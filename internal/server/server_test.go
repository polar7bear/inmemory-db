package server

import (
	"bufio"
	"net"
	"testing"
	"time"
)

// - t.Error() / t.Errorf(): 실패 기록 (테스트 계속 진행)
// - t.Fatal() / t.Fatalf(): 실패 기록 후 즉시 중단

// New(":6379")가 nil이 아닌 Server를 반환하는지
func TestServerCreation(t *testing.T) {
	// given: 바인딩할 주소
	addr := ":6379"

	// when: 서버 인스턴스 생성
	server := New(addr)

	// then: nil이 아니어야 함
	if server == nil {
		t.Fatalf("actual: %s, expected: 서버인스턴스", server)
	}
}

// 서버가 지정된 포트에서 리스닝을 시작하는지 검증
func TestServerListensOnPort(t *testing.T) {
	// given: 서버 생성 및 시작
	server := New(":6379")

	// 고루틴에서 서버 시작 (블로킹 우회)
	go server.Start()

	// 서버가 준비될 때까지 1초 슬립
	time.Sleep(time.Second)

	// when: 클라이언트로 연결 시도
	conn, err := net.Dial("tcp", "localhost:6379")

	// then
	if err != nil {
		t.Fatalf("포트 리스닝 실패")
	}

	defer conn.Close()
}

// 연결 시 "+PONG\r\n"을 응답하여야 한다.
func TestPongResponse(t *testing.T) {
	// given: 서버 생성 및 시작
	server := New(":6379")
	go server.Start()
	time.Sleep(time.Second)

	// when: 단일 클라이언트 연결 및 응답 읽기
	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		t.Fatal("연결 실패")
	}
	defer conn.Close()

	buf := make([]byte, 50)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatal("읽기 실패")
	}
	// then: 응답값과 기댓값 비교
	received := string(buf[:n])
	expected := "+PONG\r\n"

	if received != expected {
		t.Fatalf("actual: %s, expected: %s", received, expected)
	}
}

// 응답 후 서버가 연결을 종료하는지 검증 (Read()가 io.EOF를 반환하면 연결이 종료된 것)
func TestConnectionClose(t *testing.T) {
	// given: 서버 생성 및 시작
	server := New(":6379")
	go server.Start()
	time.Sleep(time.Second)

	// when: 연결 후 응답 읽기
	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		t.Fatal("연결 실패")
	}
	defer conn.Close()

	// 첫 번째 Read: PONG 응답 수신
	buf := make([]byte, 50)
	_, err = conn.Read(buf)
	if err != nil {
		t.Fatal("첫 번째 읽기 실패")
	}

	// then: 두 번째 Read 시도 -> 서버가 연결을 닫았다면 EOF 반환
	n, err := conn.Read(buf)

	// EOF 또는 읽은 바이트가 0이면 연결이 종료된 것임
	if err == nil && n > 0 {
		t.Fatalf("서버가 연결을 종료하지 않음. 추가로 읽은 데이터: %s", string(buf[:n]))
	}
}

// 클라이언트가 PING 명령어 전송 후 서버 PONG 응답하는지 검증
func TestReadPingCommand(t *testing.T) {
	// given: 서버 시작
	server := New(":6379")
	go server.Start()
	time.Sleep(time.Second)

	// when: 클라이언트가 연결 후 PING 명령어 전송하고 응답 수신
	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		t.Fatal("연결 실패")
	}
	defer conn.Close()

	conn.Write([]byte("PING\r\n"))
	reader := bufio.NewReader(conn)
	read,err := reader.ReadString('\n')
	
	if err != nil {
		t.Fatal("응답 읽기 실패")
	}

	// then: 응답값이 일치한지 검증
	if read != "+PONG\r\n" {
		t.Fatalf("잘못 된 응답 값: %s", read)
	}
}
