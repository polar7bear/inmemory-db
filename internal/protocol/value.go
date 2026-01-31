package protocol

type Value struct {
	Type  byte   // '+', '-', ':', '$', '*'
    Str   string
    Num   int
    Array []Value
}
