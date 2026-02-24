package persistence

var MagicBytes = [6]byte{'M', 'I', 'N', 'I', 'D', 'B'}

const (
	Version byte = 0x01

	TypeString byte = 0x00
	TypeList   byte = 0x01

	NoExpiry  byte = 0x00
	HasExpiry byte = 0x01

	EOF byte = 0xFF

	ChecksumSize = 4
)
