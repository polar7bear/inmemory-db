package server

import (
	"bufio"
	"inmemory-db/internal/protocol"
	"inmemory-db/internal/storage"
	"log"
	"net"
	"strconv"
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

	if err := s.store.Load("dump.rdb"); err != nil {
		log.Printf("RDB 로딩 실패: %v", err)
	} else {
		log.Println("RDB 파일 로딩 완료")
	}

	s.store.StartExpiry()
	defer s.store.StopExpiry()

	// 무한루프
	for {
		// Accept() 호출 -> 연결대기(블로킹)
		conn, err := s.listener.Accept()

		if err != nil {
			log.Print("연결 중 오류: ", err)
			continue
		} else {
			go s.handleConnection(conn)
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

		case "LPUSH":
			if len(value.Array) < 3 {
				writer.WriteError("missing argument")
			} else {
				values := make([]string, 0, len(value.Array)-2)
				for _, v := range value.Array[2:] {
					values = append(values, v.Str)
				}
				length, err := s.store.LPush(value.Array[1].Str, values...)
				if err != nil {
					writer.WriteError(err.Error())
				} else {
					writer.WriteInteger(length)
				}
			}

		case "RPUSH":
			if len(value.Array) < 3 {
				writer.WriteError("missing argument")
			} else {
				values := make([]string, 0, len(value.Array)-2)
				for _, v := range value.Array[2:] {
					values = append(values, v.Str)
				}
				length, err := s.store.RPush(value.Array[1].Str, values...)
				if err != nil {
					writer.WriteError(err.Error())
				} else {
					writer.WriteInteger(length)
				}
			}

		case "LPOP":
			if len(value.Array) < 2 {
				writer.WriteError("missing argument")
			} else {
				value, result, err := s.store.LPop(value.Array[1].Str)
				if err != nil {
					writer.WriteError(err.Error())
				} else {
					if !result {
						writer.WriteNull()
					} else {
						writer.WriteBulkString(value)
					}
				}
			}

		case "RPOP":
			if len(value.Array) < 2 {
				writer.WriteError("missing argument")
			} else {
				value, result, err := s.store.RPop(value.Array[1].Str)
				if err != nil {
					writer.WriteError(err.Error())
				} else {
					if !result {
						writer.WriteNull()
					} else {
						writer.WriteBulkString(value)
					}
				}
			}

		case "LRANGE":
			if len(value.Array) < 4 {
				writer.WriteError("missing argument")
			} else {
				start, _ := strconv.Atoi(value.Array[2].Str)
				stop, _ := strconv.Atoi(value.Array[3].Str)
				result, err := s.store.LRange(value.Array[1].Str, start, stop)
				if err != nil {
					writer.WriteError(err.Error())
				} else {
					writer.WriteArray(result)
				}
			}

		case "EXPIRE":
			if len(value.Array) < 3 {
				writer.WriteError("missing argument")
			} else {
				seconds, _ := strconv.Atoi(value.Array[2].Str)
				result := s.store.Expire(value.Array[1].Str, seconds)
				writer.WriteInteger(result)
			}

		case "TTL":
			if len(value.Array) < 2 {
				writer.WriteError("missing argument")
			} else {
				result := s.store.TTL(value.Array[1].Str)
				writer.WriteInteger(result)
			}

		case "DEL":
			if len(value.Array) < 2 {
				writer.WriteError("missing argument")
			} else {
				result := s.store.Del(value.Array[1].Str)
				writer.WriteInteger(result)
			}

		case "PERSIST":
			if len(value.Array) < 2 {
				writer.WriteError("missing argument")
			} else {
				result := s.store.Persist(value.Array[1].Str)
				writer.WriteInteger(result)
			}

		case "SAVE":
			err := s.store.Save("dump.rdb")
			if err != nil {
				writer.WriteError(err.Error())
			} else {
				writer.WriteSimpleString("OK")
			}

		default:
			writer.WriteError("unknown command")
		}
	}
}
