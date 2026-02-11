package server

import (
	"bufio"
	"fmt"
	"net"
	"sync"
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
	input := "*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$6\r\n승기\r\n"
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

func TestConcurrentClients(t *testing.T) {
	// given
	input := "*1\r\n$4\r\nPING\r\n"
	server := New(":6379")
	go server.Start()
	time.Sleep(time.Second)

	// when: 여러 클라이언트가 동시에 접속
	var wg sync.WaitGroup
	clientCount := 10

	for i := 0; i < clientCount; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			conn, _ := net.Dial("tcp", "localhost:6379")
			defer conn.Close()

			conn.Write([]byte(input))

			reader := bufio.NewReader(conn)
			response, _ := reader.ReadString('\n')

			// then: 모든 클라이언트가 PONG 응답
			if response != "+PONG\r\n" {
				t.Errorf("클라이언트 %d 응답 실패: %s", id, response)
			}
		}(i)
	}

	wg.Wait()
}

func TestConcurrentSetGet(t *testing.T) {
	// given
	server := New(":6379")
	go server.Start()
	time.Sleep(time.Second)

	var wg sync.WaitGroup

	// when: 여러 클라이언트가 동시에 SET/GET
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			conn, _ := net.Dial("tcp", "localhost:6379")
			defer conn.Close()

			key := fmt.Sprintf("key%d", id)
			value := fmt.Sprintf("value%d", id)
			keyLen := len(key)
			valueLen := len(value)

			// SET
			setCmd := fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", keyLen, key, valueLen, value)
			conn.Write([]byte(setCmd))

			reader := bufio.NewReader(conn)
			setResponse, _ := reader.ReadString('\n')

			if setResponse != "+OK\r\n" {
				t.Errorf("클라이언트 %d SET 실패: %s", id, setResponse)
				return
			}

			// GET
			getCmd := fmt.Sprintf("*2\r\n$3\r\nGET\r\n$%d\r\n%s\r\n", keyLen, key)
			conn.Write([]byte(getCmd))

			getResponse1, _ := reader.ReadString('\n')
			getResponse2, _ := reader.ReadString('\n')
			getResp := getResponse1 + getResponse2

			expected := fmt.Sprintf("$%d\r\n%s\r\n", valueLen, value)
			if getResp != expected {
				t.Errorf("클라이언트 %d GET 실패: %s, expected: %s", id, getResp, expected)
			}
		}(i)
	}

	wg.Wait()
	// then: -race로 실행 시 race가 감지되지 않아야 함
}

// ===== List 명령어 통합 테스트 =====

func TestLPushCommand(t *testing.T) {
	// given
	server := New(":6379")
	go server.Start()
	time.Sleep(time.Second)

	conn, _ := net.Dial("tcp", "localhost:6379")
	defer conn.Close()

	// when: LPUSH lp-list a b c
	conn.Write([]byte("*5\r\n$5\r\nLPUSH\r\n$7\r\nlp-list\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n"))

	// then: 길이 3 반환
	reader := bufio.NewReader(conn)
	response, _ := reader.ReadString('\n')

	if response != ":3\r\n" {
		t.Fatalf("응답: %s", response)
	}
}

func TestRPushCommand(t *testing.T) {
	// given
	server := New(":6379")
	go server.Start()
	time.Sleep(time.Second)

	conn, _ := net.Dial("tcp", "localhost:6379")
	defer conn.Close()

	// when: RPUSH rp-list x y
	conn.Write([]byte("*4\r\n$5\r\nRPUSH\r\n$7\r\nrp-list\r\n$1\r\nx\r\n$1\r\ny\r\n"))

	// then
	reader := bufio.NewReader(conn)
	response, _ := reader.ReadString('\n')

	if response != ":2\r\n" {
		t.Fatalf("응답: %s", response)
	}
}

func TestLPopCommand(t *testing.T) {
	// given: LPUSH lpop-list a b c → [c, b, a]
	server := New(":6379")
	go server.Start()
	time.Sleep(time.Second)

	conn, _ := net.Dial("tcp", "localhost:6379")
	defer conn.Close()

	reader := bufio.NewReader(conn)
	conn.Write([]byte("*5\r\n$5\r\nLPUSH\r\n$9\r\nlpop-list\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n"))
	reader.ReadString('\n') // :3 소비

	// when: LPOP lpop-list
	conn.Write([]byte("*2\r\n$4\r\nLPOP\r\n$9\r\nlpop-list\r\n"))

	// then: Head에서 제거 → c
	line1, _ := reader.ReadString('\n')
	line2, _ := reader.ReadString('\n')
	response := line1 + line2

	if response != "$1\r\nc\r\n" {
		t.Fatalf("응답: %s", response)
	}
}

func TestRPopCommand(t *testing.T) {
	// given: LPUSH rpop-list a b c → [c, b, a]
	server := New(":6379")
	go server.Start()
	time.Sleep(time.Second)

	conn, _ := net.Dial("tcp", "localhost:6379")
	defer conn.Close()

	reader := bufio.NewReader(conn)
	conn.Write([]byte("*5\r\n$5\r\nLPUSH\r\n$9\r\nrpop-list\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n"))
	reader.ReadString('\n') // :3 소비

	// when: RPOP rpop-list
	conn.Write([]byte("*2\r\n$4\r\nRPOP\r\n$9\r\nrpop-list\r\n"))

	// then: Tail에서 제거 → a
	line1, _ := reader.ReadString('\n')
	line2, _ := reader.ReadString('\n')
	response := line1 + line2

	if response != "$1\r\na\r\n" {
		t.Fatalf("응답: %s", response)
	}
}

func TestLRangeCommand(t *testing.T) {
	// given: RPUSH range-list a b c → [a, b, c]
	server := New(":6379")
	go server.Start()
	time.Sleep(time.Second)

	conn, _ := net.Dial("tcp", "localhost:6379")
	defer conn.Close()

	reader := bufio.NewReader(conn)
	conn.Write([]byte("*5\r\n$5\r\nRPUSH\r\n$10\r\nrange-list\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n"))
	reader.ReadString('\n') // :3 소비

	// when: LRANGE range-list 0 -1
	conn.Write([]byte("*4\r\n$6\r\nLRANGE\r\n$10\r\nrange-list\r\n$1\r\n0\r\n$2\r\n-1\r\n"))

	// then: *3\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n
	// 1줄(배열 헤더) + 2줄*3(bulk string) = 7줄
	var response string
	for i := 0; i < 7; i++ {
		line, _ := reader.ReadString('\n')
		response += line
	}

	expected := "*3\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n"
	if response != expected {
		t.Fatalf("응답: %q, expected: %q", response, expected)
	}
}

func TestLPopNonExistentKey(t *testing.T) {
	// given
	server := New(":6379")
	go server.Start()
	time.Sleep(time.Second)

	conn, _ := net.Dial("tcp", "localhost:6379")
	defer conn.Close()

	// when: 존재하지 않는 키에 LPOP
	conn.Write([]byte("*2\r\n$4\r\nLPOP\r\n$10\r\nno-key-pop\r\n"))

	// then: Null 응답
	reader := bufio.NewReader(conn)
	response, _ := reader.ReadString('\n')

	if response != "$-1\r\n" {
		t.Fatalf("응답: %s", response)
	}
}

func TestLRangeNonExistentKey(t *testing.T) {
	// given
	server := New(":6379")
	go server.Start()
	time.Sleep(time.Second)

	conn, _ := net.Dial("tcp", "localhost:6379")
	defer conn.Close()

	// when: 존재하지 않는 키에 LRANGE
	conn.Write([]byte("*4\r\n$6\r\nLRANGE\r\n$12\r\nno-key-range\r\n$1\r\n0\r\n$2\r\n-1\r\n"))

	// then: 빈 배열
	reader := bufio.NewReader(conn)
	response, _ := reader.ReadString('\n')

	if response != "*0\r\n" {
		t.Fatalf("응답: %s", response)
	}
}

func TestWrongTypeError(t *testing.T) {
	// given: SET으로 String 타입 저장
	server := New(":6379")
	go server.Start()
	time.Sleep(time.Second)

	conn, _ := net.Dial("tcp", "localhost:6379")
	defer conn.Close()

	reader := bufio.NewReader(conn)
	conn.Write([]byte("*3\r\n$3\r\nSET\r\n$8\r\nwt-key01\r\n$5\r\nvalue\r\n"))
	reader.ReadString('\n') // +OK 소비

	// when: 같은 키에 LPUSH 시도
	conn.Write([]byte("*3\r\n$5\r\nLPUSH\r\n$8\r\nwt-key01\r\n$1\r\na\r\n"))

	// then: WRONGTYPE 에러
	response, _ := reader.ReadString('\n')
	expected := "-ERR WRONGTYPE Operation against a key holding the wrong kind of value\r\n"
	if response != expected {
		t.Fatalf("응답: %s", response)
	}
}
