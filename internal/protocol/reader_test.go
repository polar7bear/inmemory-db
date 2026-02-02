package protocol

import (
	"bufio"
	"strings"
	"testing"
)

// Input Simple String을 읽고 파싱 후 검증
func TestReadSimpleString(t *testing.T) {
	// given
	input := "+OK\r\n"
	// 메모리에 있는 문자열을 파일이나 네트워크처럼 "읽을 수 있는" 형태로 만듦
	StrReader := strings.NewReader(input)
	bufioReader := bufio.NewReader(StrReader)

	// when: RESP Reader 생성 후 Read() 함수 실행
	reader := NewReader(bufioReader)
	value, err := reader.Read()

	if err != nil {
		t.Fatal("읽어오기 실패")
	}

	// then: 파싱 결과 검증
	if value.Type != '+' || value.Str != "OK" {
		t.Fatalf("타입: %q 문자열: %s", value.Type, value.Str)
	}
}

// 에러 응답 검증
func TestReadError(t *testing.T) {
	// given
	input := "-ERR unknown command\r\n"
	StrReader := strings.NewReader(input)
	bufioReader := bufio.NewReader(StrReader)

	// when
	reader := NewReader(bufioReader)
	value, _ := reader.Read()

	// then
	if value.Type != '-' || value.Str != "ERR unknown command" {
		t.Fatalf("타입: %q 문자열: %s", value.Type, value.Str)
	}
}

// bulk string 응답 검증
func TestReadBulkString(t *testing.T) {
	// given
	input := "$5\r\nhello\r\n"
	StrReader := strings.NewReader(input)
	bufReader := bufio.NewReader(StrReader)

	// when
	reader := NewReader(bufReader)
	value, _ := reader.Read()

	// then
	if value.Type != '$' || value.Str != "hello" {
		t.Fatalf("타입: %q 문자열: %s", value.Type, value.Str)
	}
}

func TestReadNullBulkString(t *testing.T) {
	// given
	input := "$-1\r\n"
	StrReader := strings.NewReader(input)
	bufReader := bufio.NewReader(StrReader)

	// when
	reader := NewReader(bufReader)
	value, _ := reader.Read()

	// then
	if value.Type != '$' || value.Str != "Null" {
		t.Fatalf("타입: %q 문자열: %s", value.Type, value.Str)
	}
}

func TestReadArray(t *testing.T) {
	// given
	input := "*2\r\n$4\r\nECHO\r\n$5\r\nhello\r\n"
	StrReader := strings.NewReader(input)
	bufReader := bufio.NewReader(StrReader)

	// when
	reader := NewReader(bufReader)
	value, _ := reader.Read()

	// then
	if value.Type != '*' {
		t.Fatalf("타입: %q", value.Type)
	}
	if len(value.Array) != 2 {
		t.Fatalf("배열 길이가 다름: %d", len(value.Array))
	}
	if value.Array[0].Str != "ECHO" || value.Array[1].Str != "hello" {
		t.Fatalf("배열 요소가 다릅니다.")
	}
}
