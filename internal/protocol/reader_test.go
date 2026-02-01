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
