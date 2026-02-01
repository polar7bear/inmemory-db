package protocol

import (
	"bufio"
	"strings"
	"testing"
)

func TestReadSimpleString(t *testing.T) {
	input := "+OK\r\n"
	StrReader := strings.NewReader(input)
	bufioReader := bufio.NewReader(StrReader)
	reader := NewReader(bufioReader)

	value, err := reader.Read()
	if err != nil {
		t.Fatal("읽어오기 실패")
	}

	if value.Type != '+' || value.Str != "OK" {
		t.Fatalf("타입: %q 문자열: %s", value.Type, value.Str)
	}
}
