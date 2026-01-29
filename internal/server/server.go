package server

import (
	"bufio"
	"log"
	"net"
	"strings"
)

// TCP 서버
type Server struct {
	listener net.Listener
	addr     string
}

func New(addr string) *Server {
	return &Server{addr: addr}
}

// Start는 서버를 시작하고 연결을 수신합니다.
// 이 함수는 블로킹됩니다 (무한 루프).
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)

	if err != nil {
		return err
	}

	// 서버 리스너에 저장
	s.listener = listener
	defer s.listener.Close() // 서버 종료 전 리소스 정리
	log.Printf("현재 서버가 [%s] 에서 리스닝중입니다.", s.addr)

	// 무한루프
	for {
		// Accept() 호출 -> 연결대기(블로킹)
		conn, err := s.listener.Accept()

		if err != nil {
			log.Print("연결 중 오류: ", err)
			continue
		} else {
			s.handleConnection(conn)
		}
	}

}

// 단일 클라이언트 연결을 처리합니다.
func (s *Server) handleConnection(conn net.Conn) {
	// 연결 종료 예약
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		read, err := reader.ReadString('\n')
		// EOF면 클라이언트가 연결을 끊은 것이니 루프 탈출 필요
		if err != nil {
			return
		}

		// 접미사 제거
		// 맥/리눅스 환경에서는 nc(netcat)에서 엔터를 치면 \n만 전송
		// 윈도우 환경에서는 \r\n 전송 -> OS환경마다 다름.
		// 이 때문에 그냥 TrimSpace() 사용. -> 앞뒤의 모든 공백문자 제거
		// read = strings.TrimSuffix(read, "\r\n") -> 문제되는 코드
		read = strings.TrimSpace(read)

		// SplitN(타겟 문자열, 구분자, split할 부분문자열의 개수)
		var commandList []string = strings.SplitN(read, " ", 2)
		command := strings.ToUpper(commandList[0])

		switch command {

		case "PING":
			conn.Write([]byte("+PONG\r\n"))

		case "ECHO":
			if len(commandList) > 1 {
				conn.Write([]byte("+" + commandList[1] + "\r\n"))
			} else {
				conn.Write([]byte("-ERROR missing argument\r\n"))
			}

		default:
			conn.Write([]byte("-ERROR unknown command\r\n"))
		}
	}
	// 응답 전송
	// conn.Write([]byte("+PONG\r\n"))

	// log.Printf("클라이언트가 성공적으로 연결되었습니다: %s", conn.RemoteAddr())
}
