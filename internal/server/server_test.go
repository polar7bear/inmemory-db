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
	read, err := reader.ReadString('\n')

	if err != nil {
		t.Fatal("응답 읽기 실패")
	}

	// then: 응답값이 일치한지 검증
	if read != "+PONG\r\n" {
		t.Fatalf("잘못 된 응답 값: %s", read)
	}
}

// 클라이언트가 "ECHO ..." 명령어 전송 후 서버가 정상적으로 응답하는지 검증
func TestReadEchoCommand(t *testing.T) {
	// given: 서버 시작
	server := New(":6379")
	go server.Start()
	time.Sleep(time.Second)

	// when: 클라이언트가 연결 후 echo 명령어 전송하고 응답 수신
	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		t.Fatal("연결 실패")
	}
	defer conn.Close()

	conn.Write([]byte("ECHO HELLO\r\n"))
	reader := bufio.NewReader(conn)
	read, err := reader.ReadString('\n')

	if err != nil {
		t.Fatal("응답 읽기 실패")
	}

	// then: 응답값이 일치한지 검증
	if read != "+HELLO\r\n" {
		t.Fatalf("잘못 된 응답 값: %s", read)
	}
}

// 여러 명령어를 수신 후 모든 명령어에 대하여 정상적으로 응답하는지 검증
func TestMultipleCommands(t *testing.T) {
	// given: 서버 시작
	server := New(":6379")
	go server.Start()
	time.Sleep(time.Second)

	// when: 클라이언트가 연결 후 여러 명령어를 순차적으로 전송
	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		t.Fatal("연결 실패")
	}
	defer conn.Close()

	// 첫 번째 명령어 PING
	conn.Write([]byte("PING\r\n"))
	reader := bufio.NewReader(conn)
	read, _ := reader.ReadString('\n')

	if read != "+PONG\r\n" {
		t.Fatalf("잘못 된 응답 값: %s", read)
	}

	// 두 번째 명령어 ECHO HELLO
	conn.Write([]byte("ECHO HELLO\r\n"))
	read2, _ := reader.ReadString('\n')

	if read2 != "+HELLO\r\n" {
		t.Fatalf("잘못 된 응답 값: %s", read2)
	}

	// 세 번째 명령어 PING
	conn.Write([]byte("PING\r\n"))
	read3, _ := reader.ReadString('\n')

	if read3 != "+PONG\r\n" {
		t.Fatalf("잘못 된 응답 값: %s", read3)
	}
}

// 클라이언트가 서버가 알 수 없는 명령어를 전송 후 서버가 에러를 정상적으로 응답하는지 검증
func TestInvalidCommand(t *testing.T) {
	// given: 서버 시작
	server := New(":6379")
	go server.Start()
	time.Sleep(time.Second)

	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		t.Fatal("연결 실패")
	}
	defer conn.Close()

	// when: 클라이언트가 ECHO 명령어를 보내는데 인자를 작성하지 않았을 때
	conn.Write([]byte("ECHO\r\n"))
	reader := bufio.NewReader(conn)
	read, _ := reader.ReadString('\n')

	// then: 에러 응답 검증
	if read != "-ERR missing argument\r\n" {
		t.Fatalf("잘못 된 응답 값: %s", read)
	}

	// when: 클라이언트가 서버에서 알 수 없는 명령어를 보냈을 때
	conn.Write([]byte("HELLO\r\n"))
	read2, _ := reader.ReadString('\n')

	// then: 에러 응답 검증
	if read2 != "-ERR unknown command\r\n" {
		t.Fatalf("잘못 된 응답 값: %s", read2)
	}
}

// 서버가 클라이언트로부터 수신받은 명령어에 대해서 대소문자 구분없이 모두 잘 처리하는지 검증
func TestCaseInsensitiveCommand(t *testing.T) {
	// given: 서버 시작
	server := New(":6379")
	go server.Start()
	time.Sleep(time.Second)

	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		t.Fatal("연결 실패")
	}
	defer conn.Close()

	// when: 클라이언트가 대소문자 구분없이 명령어 전송
	conn.Write([]byte("pInG\r\n"))
	reader := bufio.NewReader(conn)
	read, _ := reader.ReadString('\n')

	// then: 응답 검증
	if read != "+PONG\r\n" {
		t.Fatalf("잘못 된 응답 값: %s", read)
	}
}

func TestRespPingCommand(t *testing.T) {
	// given
	server := New(":6379")
	go server.Start()
	time.Sleep(time.Second)

	conn, _ := net.Dial("tcp", "localhost:6379")
	defer conn.Close()

	// when: RESP 형식으로 PING 명령어 전송
	conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))

	// then
	reader := bufio.NewReader(conn)
	response, _ := reader.ReadString('\n')

	if response != "+PONG\r\n" {
		t.Fatalf("응답: %s", response)
	}
}

func TestRespEchoCommand(t *testing.T) {
	// given
	server := New(":6379")
	go server.Start()
	time.Sleep(time.Second)

	conn, _ := net.Dial("tcp", "localhost:6379")
	defer conn.Close()

	// when: RESP 형식으로 PING 명령어 전송
	conn.Write([]byte("*2\r\n$4\r\nECHO\r\n$5\r\nhello\r\n"))

	// then
	reader := bufio.NewReader(conn)

	// bul string은 두 줄로 응답됨 -> "$5\r\n" + "hello\r\n"
	line1, _ := reader.ReadString('\n')
	line2, _ := reader.ReadString('\n')

	response := line1 + line2

	if response != "$5\r\nhello\r\n" {
		t.Fatalf("응답: %s", response)
	}
}