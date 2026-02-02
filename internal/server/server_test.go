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
		t.Fatalf("actual: %v, expected: 서버인스턴스", server)
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

func TestSetCommand(t *testing.T) {
	// given
	input := "*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$5\r\nredis\r\n" // 버퍼 바이트 크기 확인필요
	server := New(":6379")
	go server.Start()
	time.Sleep(time.Second)

	conn, _ := net.Dial("tcp", "localhost:6379")
	defer conn.Close()

	// when
	conn.Write([]byte(input))

	// then
	reader := bufio.NewReader(conn)
	response, _ := reader.ReadString('\n')

	if response != "+OK\r\n" {
		t.Fatalf("응답: %s", response)
	}
}

func TestGetCommand(t *testing.T) {
	// given
	input := "*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$5\r\nredis\r\n"
	get := "*2\r\n$3\r\nGET\r\n$4\r\nname\r\n"
	server := New(":6379")
	go server.Start()
	time.Sleep(time.Second)

	conn, _ := net.Dial("tcp", "localhost:6379")
	defer conn.Close()

	// when
	conn.Write([]byte(input))
	conn.Write([]byte(get))

	// then
	reader := bufio.NewReader(conn)
	reader.ReadString('\n')
	response2, _ := reader.ReadString('\n')
	response3, _ := reader.ReadString('\n')

	getResponse := response2 + response3

	if getResponse != "$5\r\nredis\r\n" {
		t.Fatalf("응답: %s", getResponse)
	}
}

func TestGetNonExistent(t *testing.T) {
	// given
	input := "*2\r\n$3\r\nGET\r\n$4\r\neman\r\n"
	server := New(":6379")
	go server.Start()
	time.Sleep(time.Second)

	conn, _ := net.Dial("tcp", "localhost:6379")
	defer conn.Close()

	// when
	conn.Write([]byte(input))

	// then
	reader := bufio.NewReader(conn)
	response, _ := reader.ReadString('\n')

	if response != "$-1\r\n" {
		t.Fatalf("응답: %s", response)
	}
}