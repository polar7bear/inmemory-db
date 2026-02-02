package server

import (
	"bufio"
	"inmemory-db/internal/protocol"
	"inmemory-db/internal/storage"
	"log"
	"net"
	"strings"
)

// TCP 서버
type Server struct {
	listener net.Listener
	addr     string
	store    *storage.Store
}

func New(addr string) *Server {
	return &Server{
		addr:  addr,
		store: storage.New(),
	}
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

	bufReader := bufio.NewReader(conn)
	reader := protocol.NewReader(bufReader)
	writer := protocol.NewWriter(conn)

	for {
		value, err := reader.Read()
		// EOF면 클라이언트가 연결을 끊은 것이니 루프 탈출 필요
		if err != nil {
			return
		}

		command := strings.ToUpper(value.Array[0].Str)

		switch command {

		case "PING":
			writer.WriteSimpleString("PONG")

		case "ECHO":
			if len(value.Array) > 1 {
				writer.WriteBulkString(value.Array[1].Str)
			} else {
				writer.WriteError("missing argument")
			}
	
		case "SET":
			key := value.Array[1].Str
			value := value.Array[2].Str
			s.store.Set(key, value)

			writer.WriteSimpleString("OK")

		case "GET":
			key := value.Array[1].Str
			result, exist := s.store.Get(key)
			if exist {
				writer.WriteBulkString(result)
			} else {
				writer.WriteNull()
			}

		default:
			writer.WriteError("unknown command")
		}
	}
}
