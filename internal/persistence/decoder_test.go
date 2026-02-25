package persistence

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Encoder로 쓴 데이터를 Decoder로 읽어서 검증하는 헬퍼
func encodeToBytes(t *testing.T, fn func(enc *Encoder)) []byte {
	t.Helper()
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	fn(enc)
	enc.Flush()
	return buf.Bytes()
}

func TestReadHeader(t *testing.T) {
	// given: Encoder로 작성한 헤더
	data := encodeToBytes(t, func(enc *Encoder) {
		enc.WriteHeader()
	})
	decoder := NewDecoder(bytes.NewReader(data))

	// when
	err := decoder.ReadHeader()

	// then
	if err != nil {
		t.Fatalf("에러 발생: %v", err)
	}
}

func TestReadHeader_InvalidMagicBytes(t *testing.T) {
	// given: 잘못된 Magic Bytes
	data := []byte{'B', 'A', 'D', 'B', 'A', 'D', 0x01}
	decoder := NewDecoder(bytes.NewReader(data))

	// when
	err := decoder.ReadHeader()

	// then
	if err == nil {
		t.Fatal("에러가 발생해야 합니다")
	}
}

func TestReadHeader_InvalidVersion(t *testing.T) {
	// given: 올바른 Magic Bytes + 잘못된 버전
	data := []byte{'M', 'I', 'N', 'I', 'D', 'B', 0x99}
	decoder := NewDecoder(bytes.NewReader(data))

	// when
	err := decoder.ReadHeader()

	// then
	if err == nil {
		t.Fatal("에러가 발생해야 합니다")
	}
}

func TestReadStringEntry(t *testing.T) {
	// given: Header + String("hello", "world") + EOF
	data := encodeToBytes(t, func(enc *Encoder) {
		enc.WriteHeader()
		enc.WriteStringEntry("hello", "world", nil)
		enc.WriteEOF()
	})
	decoder := NewDecoder(bytes.NewReader(data))
	decoder.ReadHeader()

	// when
	entry, err := decoder.ReadEntry()

	// then
	if err != nil {
		t.Fatalf("에러 발생: %v", err)
	}
	if entry.Type != TypeString {
		t.Fatalf("Type: 0x%02x, expected: 0x%02x", entry.Type, TypeString)
	}
	if entry.Key != "hello" {
		t.Fatalf("Key: %s, expected: hello", entry.Key)
	}
	if entry.Value != "world" {
		t.Fatalf("Value: %s, expected: world", entry.Value)
	}
	if entry.ExpireAt != nil {
		t.Fatalf("ExpireAt이 nil이어야 합니다")
	}
}

func TestReadListEntry(t *testing.T) {
	// given: Header + List("mylist", ["a", "b", "c"]) + EOF
	data := encodeToBytes(t, func(enc *Encoder) {
		enc.WriteHeader()
		enc.WriteListEntry("mylist", []string{"a", "b", "c"}, nil)
		enc.WriteEOF()
	})
	decoder := NewDecoder(bytes.NewReader(data))
	decoder.ReadHeader()

	// when
	entry, err := decoder.ReadEntry()

	// then
	if err != nil {
		t.Fatalf("에러 발생: %v", err)
	}
	if entry.Type != TypeList {
		t.Fatalf("Type: 0x%02x, expected: 0x%02x", entry.Type, TypeList)
	}
	if entry.Key != "mylist" {
		t.Fatalf("Key: %s, expected: mylist", entry.Key)
	}
	if len(entry.Values) != 3 {
		t.Fatalf("Values 길이: %d, expected: 3", len(entry.Values))
	}
	expected := []string{"a", "b", "c"}
	for i, v := range entry.Values {
		if v != expected[i] {
			t.Fatalf("Values[%d]: %s, expected: %s", i, v, expected[i])
		}
	}
}

func TestReadEntryWithTTL(t *testing.T) {
	// given: TTL이 설정된 String 엔트리
	expireAt := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	data := encodeToBytes(t, func(enc *Encoder) {
		enc.WriteHeader()
		enc.WriteStringEntry("session", "token", &expireAt)
		enc.WriteEOF()
	})
	decoder := NewDecoder(bytes.NewReader(data))
	decoder.ReadHeader()

	// when
	entry, err := decoder.ReadEntry()

	// then
	if err != nil {
		t.Fatalf("에러 발생: %v", err)
	}
	if entry.ExpireAt == nil {
		t.Fatal("ExpireAt이 nil이면 안 됩니다")
	}
	if entry.ExpireAt.UnixMilli() != expireAt.UnixMilli() {
		t.Fatalf("ExpireAt: %d, expected: %d", entry.ExpireAt.UnixMilli(), expireAt.UnixMilli())
	}
}

func TestReadEntry_EOF(t *testing.T) {
	// given: Header + EOF만 있는 파일
	data := encodeToBytes(t, func(enc *Encoder) {
		enc.WriteHeader()
		enc.WriteEOF()
	})
	decoder := NewDecoder(bytes.NewReader(data))
	decoder.ReadHeader()

	// when
	entry, err := decoder.ReadEntry()

	// then: EOF면 nil, nil 반환
	if err != nil {
		t.Fatalf("에러 발생: %v", err)
	}
	if entry != nil {
		t.Fatal("EOF에서 entry는 nil이어야 합니다")
	}
}

func TestReadMultipleEntries(t *testing.T) {
	// given: String 2개 + List 1개
	data := encodeToBytes(t, func(enc *Encoder) {
		enc.WriteHeader()
		enc.WriteStringEntry("k1", "v1", nil)
		enc.WriteStringEntry("k2", "v2", nil)
		enc.WriteListEntry("l1", []string{"x", "y"}, nil)
		enc.WriteEOF()
	})
	decoder := NewDecoder(bytes.NewReader(data))
	decoder.ReadHeader()

	// when: 모든 엔트리 읽기
	entries := make([]*DecodedEntry, 0)
	for {
		entry, err := decoder.ReadEntry()
		if err != nil {
			t.Fatalf("에러 발생: %v", err)
		}
		if entry == nil {
			break
		}
		entries = append(entries, entry)
	}

	// then
	if len(entries) != 3 {
		t.Fatalf("엔트리 수: %d, expected: 3", len(entries))
	}
	if entries[0].Key != "k1" || entries[0].Type != TypeString {
		t.Fatalf("entries[0]: key=%s, type=0x%02x", entries[0].Key, entries[0].Type)
	}
	if entries[1].Key != "k2" || entries[1].Type != TypeString {
		t.Fatalf("entries[1]: key=%s, type=0x%02x", entries[1].Key, entries[1].Type)
	}
	if entries[2].Key != "l1" || entries[2].Type != TypeList {
		t.Fatalf("entries[2]: key=%s, type=0x%02x", entries[2].Key, entries[2].Type)
	}
}

func TestReadEntry_UnknownType(t *testing.T) {
	// given: 알 수 없는 타입 바이트 (0xAA) + 더미 키
	buf := []byte{0xAA, 0x00, 0x00, 0x00, 0x01, 'x'}
	decoder := NewDecoder(bytes.NewReader(buf))

	// when
	_, err := decoder.ReadEntry()

	// then
	if err == nil {
		t.Fatal("알 수 없는 타입에 대해 에러가 발생해야 합니다")
	}
}

// ========== Checksum 테스트 ==========

func TestWriteAndVerifyChecksum(t *testing.T) {
	// given: Encoder로 파일 작성 (Header + Entry + EOF + Checksum)
	path := filepath.Join(t.TempDir(), "checksum.rdb")
	file, _ := os.Create(path)
	enc := NewEncoder(file)
	enc.WriteHeader()
	enc.WriteStringEntry("key", "value", nil)
	enc.WriteEOF()
	enc.WriteChecksum()
	enc.Flush()
	file.Close()

	// when
	err := VerifyChecksum(path)

	// then
	if err != nil {
		t.Fatalf("체크섬 검증 실패: %v", err)
	}
}

func TestVerifyChecksum_Corrupted(t *testing.T) {
	// given: 정상 파일 작성 후 중간 바이트 변조
	path := filepath.Join(t.TempDir(), "corrupted.rdb")
	file, _ := os.Create(path)
	enc := NewEncoder(file)
	enc.WriteHeader()
	enc.WriteStringEntry("key", "value", nil)
	enc.WriteEOF()
	enc.WriteChecksum()
	enc.Flush()
	file.Close()

	// 파일 변조: 10번째 바이트를 뒤집는다
	data, _ := os.ReadFile(path)
	data[10] ^= 0xFF
	os.WriteFile(path, data, 0644)

	// when
	err := VerifyChecksum(path)

	// then
	if err == nil {
		t.Fatal("변조된 파일에 대해 에러가 발생해야 합니다")
	}
}

func TestVerifyChecksum_TooSmall(t *testing.T) {
	// given: 4바이트 미만 파일
	path := filepath.Join(t.TempDir(), "tiny.rdb")
	os.WriteFile(path, []byte{0x01, 0x02}, 0644)

	// when
	err := VerifyChecksum(path)

	// then
	if err == nil {
		t.Fatal("너무 작은 파일에 대해 에러가 발생해야 합니다")
	}
}

func TestVerifyChecksum_ManualCheck(t *testing.T) {
	// given: 수동으로 CRC32 계산 후 비교
	path := filepath.Join(t.TempDir(), "manual.rdb")
	file, _ := os.Create(path)
	enc := NewEncoder(file)
	enc.WriteHeader()
	enc.WriteEOF()
	enc.WriteChecksum()
	enc.Flush()
	file.Close()

	// when: 파일을 직접 읽어서 CRC32 검증
	data, _ := os.ReadFile(path)
	content := data[:len(data)-ChecksumSize]
	stored := binary.BigEndian.Uint32(data[len(data)-ChecksumSize:])
	computed := crc32.ChecksumIEEE(content)

	// then
	if stored != computed {
		t.Fatalf("stored=0x%08x, computed=0x%08x", stored, computed)
	}
}
