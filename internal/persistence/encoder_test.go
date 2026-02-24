package persistence

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"
)

func TestWriteHeader(t *testing.T) {
	// given
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	// when
	err := enc.WriteHeader()
	enc.Flush()

	// then
	if err != nil {
		t.Fatalf("에러 발생: %v", err)
	}
	data := buf.Bytes()
	if len(data) != 7 {
		t.Fatalf("헤더 길이: %d, expected: 7", len(data))
	}
	if string(data[:6]) != "MINIDB" {
		t.Fatalf("Magic bytes: %s, expected: MINIDB", string(data[:6]))
	}
	if data[6] != 0x01 {
		t.Fatalf("Version: 0x%02x, expected: 0x01", data[6])
	}
}

func TestWriteStringEntry(t *testing.T) {
	// given
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	// when: key="hi", value="go", TTL 없음
	err := enc.WriteStringEntry("hi", "go", nil)
	enc.Flush()

	// then
	if err != nil {
		t.Fatalf("에러 발생: %v", err)
	}
	data := buf.Bytes()

	// 총 길이: Type(1) + KeyLen(4) + Key(2) + ValLen(4) + Val(2) + NoExpiry(1) = 14
	if len(data) != 14 {
		t.Fatalf("총 길이: %d, expected: 14", len(data))
	}

	// Type: 0x00
	if data[0] != TypeString {
		t.Fatalf("Type: 0x%02x, expected: 0x00", data[0])
	}

	// Key length: 2
	keyLen := binary.BigEndian.Uint32(data[1:5])
	if keyLen != 2 {
		t.Fatalf("Key length: %d, expected: 2", keyLen)
	}

	// Key: "hi"
	if string(data[5:7]) != "hi" {
		t.Fatalf("Key: %s, expected: hi", string(data[5:7]))
	}

	// Value length: 2
	valLen := binary.BigEndian.Uint32(data[7:11])
	if valLen != 2 {
		t.Fatalf("Value length: %d, expected: 2", valLen)
	}

	// Value: "go"
	if string(data[11:13]) != "go" {
		t.Fatalf("Value: %s, expected: go", string(data[11:13]))
	}

	// No expiry: 0x00
	if data[13] != NoExpiry {
		t.Fatalf("Expiry: 0x%02x, expected: 0x00", data[13])
	}
}

func TestWriteListEntry(t *testing.T) {
	// given
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	// when: key="l", values=["a", "b"], TTL 없음
	err := enc.WriteListEntry("l", []string{"a", "b"}, nil)
	enc.Flush()

	// then
	if err != nil {
		t.Fatalf("에러 발생: %v", err)
	}
	data := buf.Bytes()

	// 총 길이: Type(1) + KeyLen(4) + Key(1) + Count(4) + Elem1Len(4) + Elem1(1) + Elem2Len(4) + Elem2(1) + NoExpiry(1) = 21
	if len(data) != 21 {
		t.Fatalf("총 길이: %d, expected: 21", len(data))
	}

	// Type: 0x01
	if data[0] != TypeList {
		t.Fatalf("Type: 0x%02x, expected: 0x01", data[0])
	}

	// Key length: 1
	keyLen := binary.BigEndian.Uint32(data[1:5])
	if keyLen != 1 {
		t.Fatalf("Key length: %d, expected: 1", keyLen)
	}

	// Key: "l"
	if string(data[5:6]) != "l" {
		t.Fatalf("Key: %s, expected: l", string(data[5:6]))
	}

	// Element count: 2
	elemCount := binary.BigEndian.Uint32(data[6:10])
	if elemCount != 2 {
		t.Fatalf("Element count: %d, expected: 2", elemCount)
	}

	// Element 1: length=1, value="a"
	elem1Len := binary.BigEndian.Uint32(data[10:14])
	if elem1Len != 1 {
		t.Fatalf("Element 1 length: %d, expected: 1", elem1Len)
	}
	if string(data[14:15]) != "a" {
		t.Fatalf("Element 1: %s, expected: a", string(data[14:15]))
	}

	// Element 2: length=1, value="b"
	elem2Len := binary.BigEndian.Uint32(data[15:19])
	if elem2Len != 1 {
		t.Fatalf("Element 2 length: %d, expected: 1", elem2Len)
	}
	if string(data[19:20]) != "b" {
		t.Fatalf("Element 2: %s, expected: b", string(data[19:20]))
	}

	// No expiry: 0x00
	if data[20] != NoExpiry {
		t.Fatalf("Expiry: 0x%02x, expected: 0x00", data[20])
	}
}

func TestWriteExpiry_NoTTL(t *testing.T) {
	// given
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	// when
	err := enc.writeExpiry(nil)
	enc.Flush()

	// then
	if err != nil {
		t.Fatalf("에러 발생: %v", err)
	}
	data := buf.Bytes()
	if len(data) != 1 {
		t.Fatalf("길이: %d, expected: 1", len(data))
	}
	if data[0] != NoExpiry {
		t.Fatalf("Expiry: 0x%02x, expected: 0x00", data[0])
	}
}

func TestWriteExpiry_WithTTL(t *testing.T) {
	// given
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	expireAt := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	// when
	err := enc.writeExpiry(&expireAt)
	enc.Flush()

	// then: HasExpiry(1) + UnixMilli(8) = 9바이트
	if err != nil {
		t.Fatalf("에러 발생: %v", err)
	}
	data := buf.Bytes()
	if len(data) != 9 {
		t.Fatalf("길이: %d, expected: 9", len(data))
	}
	if data[0] != HasExpiry {
		t.Fatalf("Expiry marker: 0x%02x, expected: 0x01", data[0])
	}
	ms := int64(binary.BigEndian.Uint64(data[1:9]))
	if ms != expireAt.UnixMilli() {
		t.Fatalf("Unix millis: %d, expected: %d", ms, expireAt.UnixMilli())
	}
}

func TestWriteEOF(t *testing.T) {
	// given
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	// when
	err := enc.WriteEOF()
	enc.Flush()

	// then
	if err != nil {
		t.Fatalf("에러 발생: %v", err)
	}
	data := buf.Bytes()
	if len(data) != 1 {
		t.Fatalf("길이: %d, expected: 1", len(data))
	}
	if data[0] != EOF {
		t.Fatalf("EOF: 0x%02x, expected: 0xFF", data[0])
	}
}

func TestEncodeFullFile(t *testing.T) {
	// given
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	expireAt := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	// when: Header + String 엔트리 + List 엔트리(TTL 포함) + EOF
	enc.WriteHeader()
	enc.WriteStringEntry("name", "go", nil)
	enc.WriteListEntry("list", []string{"a", "b"}, &expireAt)
	enc.WriteEOF()
	enc.Flush()

	// then
	data := buf.Bytes()

	// "MINIDB"로 시작
	if string(data[:6]) != "MINIDB" {
		t.Fatalf("Magic bytes: %s", string(data[:6]))
	}

	// 0xFF로 끝남
	if data[len(data)-1] != EOF {
		t.Fatalf("마지막 바이트: 0x%02x, expected: 0xFF", data[len(data)-1])
	}

	// Header(7) + String("name"=4,"go"=2: 1+4+4+4+2+1=16) + List("list"=4,["a","b"]+TTL: 1+4+4+4+4+1+4+1+1+8=32) + EOF(1) = 56
	if len(data) != 56 {
		t.Fatalf("총 길이: %d, expected: 56", len(data))
	}
}
