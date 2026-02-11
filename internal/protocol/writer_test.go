package protocol

import (
	"bytes"
	"testing"
)

func TestWriteSimpleString(t *testing.T) {
	// given: 출력을 담을 버퍼 생성 후 Writer 생성
	var buf bytes.Buffer
	writer := NewWriter(&buf)

	// when: 버퍼에 쓰기
	writer.WriteSimpleString("OK")

	// then: 검증
	if buf.String() != "+OK\r\n" {
		t.Fatalf("문자열이 다릅니다: %s", buf.String())
	}
}

func TestWriteError(t *testing.T) {
	// given
	var buf bytes.Buffer
	writer := NewWriter(&buf)

	// when
	writer.WriteError("unknown command")

	// then
	if buf.String() != "-ERR unknown command\r\n" {
		t.Fatalf("문자열이 다릅니다: %s", buf.String())
	}
}

func TestWriteBulkString(t *testing.T) {
	// given
	var buf bytes.Buffer
	writer := NewWriter(&buf)

	// when
	writer.WriteBulkString("hello")

	// then
	if buf.String() != "$5\r\nhello\r\n" {
		t.Fatalf("문자열이 다릅니다: %s", buf.String())
	}
}
func TestWriteNull(t *testing.T) {
	// given
	var buf bytes.Buffer
	writer := NewWriter(&buf)

	// when
	writer.WriteNull()

	// then
	if buf.String() != "$-1\r\n" {
		t.Fatalf("문자열이 다릅니다: %s", buf.String())
	}
}

func TestWriteInteger(t *testing.T) {
	// given
	var buf bytes.Buffer
	writer := NewWriter(&buf)

	// when
	writer.WriteInteger(15)

	// then
	if buf.String() != ":15\r\n" {
		t.Fatalf("문자열이 다릅니다: %s", buf.String())
	}
}

func TestWriteArray(t *testing.T) {
	// given
	var buf bytes.Buffer
	writer := NewWriter(&buf)

	// when
	writer.WriteArray([]string{"hello", "world"})

	// then
	expected := "*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n"
	if buf.String() != expected {
		t.Fatalf("문자열이 다릅니다.\nactual:   %q\nexpected: %q", buf.String(), expected)
	}
}

func TestWriteArray_Empty(t *testing.T) {
	// given
	var buf bytes.Buffer
	writer := NewWriter(&buf)

	// when
	writer.WriteArray([]string{})

	// then
	if buf.String() != "*0\r\n" {
		t.Fatalf("문자열이 다릅니다: %s", buf.String())
	}
}
