package persistence

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

type Decoder struct {
	r *bufio.Reader
}

// 디코딩한 엔트리를 담는 구조체
type DecodedEntry struct {
	Type     byte
	Key      string
	Value    string
	Values   []string
	ExpireAt *time.Time
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{bufio.NewReader(r)}
}

func (d *Decoder) ReadHeader() error {
	magicByte, err := d.readBytes(6)
	if err != nil {
		return err
	}

	if !bytes.Equal(magicByte, MagicBytes[:]) {
		return fmt.Errorf("invalid magic bytes")
	}

	version, err := d.readBytes(1)

	if version[0] != Version {
		return err
	}
	return nil
}

func (d *Decoder) ReadEntry() (*DecodedEntry, error) {
	typeBuf, err := d.readBytes(1)
	if err != nil {
		return nil, err
	}

	if typeBuf[0] == EOF {
		return nil, nil
	}

	key, err := d.readString()
	if err != nil {
		return nil, err
	}

	entry := &DecodedEntry{Type: typeBuf[0], Key: key}

	switch typeBuf[0] {
	case TypeString:
		value, err := d.readString()
		if err != nil {
			return nil, err
		}
		entry.Value = value

	case TypeList:
		count, err := d.readUint32()
		if err != nil {
			return nil, err
		}
		values := make([]string, 0, count)
		for i := 0; i < int(count); i++ {
			v, err := d.readString()
			if err != nil {
				return nil, err
			}
			values = append(values, v)
		}
		entry.Values = values

	default:
		return nil, fmt.Errorf("unknown entry type: 0x%02x", typeBuf[0])
	}

	expireAt, err := d.readExpiry()
	if err != nil {
		return nil, err
	}
	entry.ExpireAt = expireAt

	return entry, nil
}

// ========== 헬퍼 메서드 ==========
func (d *Decoder) readBytes(n int) ([]byte, error) {
	buf := make([]byte, n)
	_, err := io.ReadFull(d.r, buf)
	return buf, err
}

func (d *Decoder) readUint32() (uint32, error) {
	buf, err := d.readBytes(4)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(buf), nil
}

func (d *Decoder) readInt64() (int64, error) {
	buf, err := d.readBytes(8)
	if err != nil {
		return 0, err
	}
	return int64(binary.BigEndian.Uint64(buf)), nil
}

func (d *Decoder) readString() (string, error) {
	length, err := d.readUint32()
	if err != nil {
		return "", err
	}
	buf, err := d.readBytes(int(length))
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

func (d *Decoder) readExpiry() (*time.Time, error) {
	buf, err := d.readBytes(1)
	if err != nil {
		return nil, err
	}

	if buf[0] == NoExpiry {
		return nil, nil
	}

	ms, err := d.readInt64()
	if err != nil {
		return nil, err
	}

	t := time.UnixMilli(ms)
	return &t, nil
}
