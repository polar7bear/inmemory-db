package persistence

import (
	"bufio"
	"encoding/binary"
	"hash"
	"hash/crc32"
	"io"
	"time"
)

type Encoder struct {
	w    *bufio.Writer
	hash hash.Hash32
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w:    bufio.NewWriter(w),
		hash: crc32.NewIEEE(),
	}
}

// 파일 헤더를 쓴다. Magic bytes "MINIDB" + Version 0x01 = 총 7바이트.
func (e *Encoder) WriteHeader() error {
	if err := e.writeBytes(MagicBytes[:]); err != nil {
		return err
	}
	return e.writeBytes([]byte{Version})
}

// String 타입 엔트리를 쓴다.
// [Type 0x00] [Key] [Value] [TTL]
func (e *Encoder) WriteStringEntry(key, value string, expireAt *time.Time) error {
	if err := e.writeBytes([]byte{TypeString}); err != nil {
		return err
	}
	if err := e.writeString(key); err != nil {
		return err
	}
	if err := e.writeString(value); err != nil {
		return err
	}
	return e.writeExpiry(expireAt)
}

// List 타입 엔트리를 쓴다.
// [Type 0x01] [Key] [ElementCount] [Element1] [Element2] ... [TTL]
func (e *Encoder) WriteListEntry(key string, values []string, expireAt *time.Time) error {
	if err := e.writeBytes([]byte{TypeList}); err != nil {
		return err
	}
	if err := e.writeString(key); err != nil {
		return err
	}
	if err := e.writeUint32(uint32(len(values))); err != nil {
		return err
	}
	for _, v := range values {
		if err := e.writeString(v); err != nil {
			return err
		}
	}
	return e.writeExpiry(expireAt)
}

// EOF 마커를 쓴다. 파일의 끝을 명시적으로 표시
func (e *Encoder) WriteEOF() error {
	return e.writeBytes([]byte{EOF})
}

// 버퍼에 남아있는 데이터를 실제 Writer에 내보낸다.
// 이걸 호출하지 않으면 마지막 데이터가 파일에 기록되지 않는다
func (e *Encoder) Flush() error {
	return e.w.Flush()
}

func (e *Encoder) WriteChecksum() error {
	checksum := e.hash.Sum32()
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, checksum)
	_, err := e.w.Write(buf)
	return err
}

// ========== 헬퍼 메서드 ==========

func (e *Encoder) writeBytes(data []byte) error {
	_, err := e.w.Write(data)
	if err == nil {
		e.hash.Write(data)
	}
	return err
}

// uint32 값을 Big Endian 4바이트로 변환해서 쓴다.
// 문자열 길이, 리스트 요소 개수 등을 기록할 때 사용
func (e *Encoder) writeUint32(n uint32) error {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, n)
	return e.writeBytes(buf)
}

// int64 값을 Big Endian 8바이트로 변환해서 쓴다.
// TTL의 Unix 밀리초 타임스탬프를 기록할 때 사용
func (e *Encoder) writeInt64(n int64) error {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(n))
	return e.writeBytes(buf)
}

// Length-Prefixed 문자열을 쓴다.
// [4바이트 길이] + [문자열 바이트] 형태
// 읽는 쪽에서 "앞 4바이트를 읽으면 뒤에 몇 바이트가 오는지 알 수 있다"는 것이 핵심
func (e *Encoder) writeString(s string) error {
	if err := e.writeUint32(uint32(len(s))); err != nil {
		return err
	}
	return e.writeBytes([]byte(s))
}

// TTL 정보를 쓴다
// expireAt이 nil이면 0x00 한 바이트만 쓴다 (만료 없음)
// nil이 아니면 0x01 + 8바이트 Unix 밀리초를 쓴다.
func (e *Encoder) writeExpiry(expireAt *time.Time) error {
	if expireAt == nil {
		return e.writeBytes([]byte{NoExpiry})
	}
	if err := e.writeBytes([]byte{HasExpiry}); err != nil {
		return err
	}
	return e.writeInt64(expireAt.UnixMilli())
}
