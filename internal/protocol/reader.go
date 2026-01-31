package protocol

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"strings"
)

type Reader struct {
	reader *bufio.Reader
}

func NewReader(rd *bufio.Reader) *Reader {
	return &Reader{rd}
}

func (r *Reader) Read() (Value, error) {
	typeByte, err := r.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	str, _ := r.reader.ReadString('\n')

	switch typeByte {
	case '+':
		return Value{
			Type: '+',
			Str:  strings.TrimSpace(str),
		}, nil

	case '-':
		return Value{
			Type: '-',
			Str:  strings.TrimSpace(str),
		}, nil

	case ':':
		parse, _ := strconv.Atoi(strings.TrimSpace(str))
		return Value{
			Type: ':',
			Num:  parse,
		}, nil

	case '$':
		length, _ := strconv.Atoi(strings.TrimSpace(str)) // 데이터 길이 가져오기

		if length == -1 {
			return Value{Type: '$', Str: "Null"}, nil
		}
		buf := make([]byte, length) // 해당 길이만큼 바이트 생성
		io.ReadFull(r.reader, buf)  // 정확히 length만큼의 바이트 읽기
		r.reader.ReadString('\n') // 마지막 \r\n 소비만 하고 버림
		return Value{
			Type: '$',
			Str:  string(buf),
		}, nil

	case '*':
		elementCount, _ := strconv.Atoi(strings.TrimSpace(str))
		arr := make([]Value, elementCount)

		for i := 0; i < len(arr); i++ {
			arr[i], _ = r.Read() // 재귀 호출
		}

		return Value{
			Type:  '*',
			Array: arr,
		}, nil

	default:
		return Value{}, errors.New("알 수 없는 타입입니다.")
	}
}
