package protocol

import (
	"io"
	"strconv"
)

type Writer struct {
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{w}
}

func (w *Writer) WriteSimpleString(s string) error {
	w.writer.Write([]byte("+" + s + "\r\n"))
	return nil
}

func (w *Writer) WriteError(s string) error {
	w.writer.Write([]byte("-ERR " + s + "\r\n"))
	return nil
}

func (w *Writer) WriteBulkString(s string) error {
	length := strconv.Itoa(len(s))
	w.writer.Write([]byte("$" + length + "\r\n" + s + "\r\n"))
	return nil
}

func (w *Writer) WriteNull() error {
	w.writer.Write([]byte("$-1\r\n"))
	return nil
}

// RESP Integer: ":15\r\n"
func (w *Writer) WriteInteger(n int) error {
	w.writer.Write([]byte(":" + strconv.Itoa(n) + "\r\n"))
	return nil
}

// RESP Array: "*3\r\n$5\r\nhello\r\n$5\r\nworld\r\n$3\r\nfoo\r\n"
// 빈 배열이면 "*0\r\n"
func (w *Writer) WriteArray(values []string) error {
	length := len(values)
	strLen := strconv.Itoa(length)

	w.writer.Write([]byte("*" + strLen + "\r\n"))

	for _, v := range values {
		w.WriteBulkString(v)
	}
	return nil
}
