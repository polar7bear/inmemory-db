package storage

type Node struct {
	Value string
	Prev  *Node
	Next  *Node
}

type List struct {
	Head   *Node
	Tail   *Node
	Length int
}

func NewList() *List {
	return &List{}
}

// 앞쪽 삽입 - O(1)
func (l *List) LPush(value string) {
	newNode := &Node{Value: value}

	if l.Head != nil {
		newNode.Next = l.Head
		l.Head.Prev = newNode
	} else {
		l.Tail = newNode
	}

	l.Head = newNode
	l.Length++
}

// 뒤쪽 삽입 — O(1)
func (l *List) RPush(value string) {
	newNode := &Node{Value: value}

	if l.Tail != nil {
		newNode.Prev = l.Tail
		l.Tail.Next = newNode
	} else {
		l.Head = newNode
	}

	l.Tail = newNode
	l.Length++
}

// 앞쪽 삭제 — O(1). 빈 리스트면 ("", false) 반환
func (l *List) LPop() (string, bool) {
	if l.Head == nil {
		return "", false
	}

	value := l.Head.Value
	l.Head = l.Head.Next

	if l.Head != nil {
		l.Head.Prev = nil
	} else {
		l.Tail = nil
	}

	l.Length--
	return value, true
}

// 뒤쪽 삭제 — O(1). 빈 리스트면 ("", false) 반환
func (l *List) RPop() (string, bool) {
	if l.Tail == nil {
		return "", false
	}

	value := l.Tail.Value
	l.Tail = l.Tail.Prev

	if l.Tail != nil {
		l.Tail.Next = nil
	} else {
		l.Head = nil
	}

	l.Length--
	return value, true
}

// 범위 조회. 음수 인덱스 지원 (-1 = 마지막, -2 = 마지막에서 두 번째)
func (l *List) Range(start, stop int) []string {
	if start < 0 {
		start = l.Length + start
	}
	if stop < 0 {
		stop = l.Length + stop
	}

	if start < 0 {
		start = 0
	}
	if stop > l.Length-1 {
		stop = l.Length - 1
	}

	if start > stop {
		return make([]string, 0)
	}

	current := l.Head
	for i := 0; i < start; i++ {
		current = current.Next
	}

	result := make([]string, 0, stop-start+1)
	for i := start; i <= stop; i++ {
		result = append(result, current.Value)
		current = current.Next
	}

	return result
}
