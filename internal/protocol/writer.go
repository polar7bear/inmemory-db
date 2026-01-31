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