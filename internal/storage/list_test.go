package storage

import "testing"

func TestLPush_SingleElement(t *testing.T) {
	// given
	list := NewList()

	// when
	list.LPush("a")

	// then
	if list.Length != 1 {
		t.Fatalf("길이가 다릅니다. actual: %d, expected: 1", list.Length)
	}
	if list.Head.Value != "a" {
		t.Fatalf("Head 값이 다릅니다. actual: %s, expected: a", list.Head.Value)
	}
	if list.Tail.Value != "a" {
		t.Fatalf("Tail 값이 다릅니다. actual: %s, expected: a", list.Tail.Value)
	}
}

func TestLPush_MultipleElements(t *testing.T) {
	// given
	list := NewList()

	// when: a, b, c 순서로 LPush → [c, b, a]
	list.LPush("a")
	list.LPush("b")
	list.LPush("c")

	// then
	if list.Length != 3 {
		t.Fatalf("길이가 다릅니다. actual: %d, expected: 3", list.Length)
	}
	if list.Head.Value != "c" {
		t.Fatalf("Head 값이 다릅니다. actual: %s, expected: c", list.Head.Value)
	}
	if list.Tail.Value != "a" {
		t.Fatalf("Tail 값이 다릅니다. actual: %s, expected: a", list.Tail.Value)
	}
	// 포인터 연결 검증: c → b → a
	if list.Head.Next.Value != "b" {
		t.Fatalf("Head.Next 값이 다릅니다. actual: %s, expected: b", list.Head.Next.Value)
	}
	if list.Tail.Prev.Value != "b" {
		t.Fatalf("Tail.Prev 값이 다릅니다. actual: %s, expected: b", list.Tail.Prev.Value)
	}
}

func TestRPush_SingleElement(t *testing.T) {
	// given
	list := NewList()

	// when
	list.RPush("a")

	// then
	if list.Length != 1 {
		t.Fatalf("길이가 다릅니다. actual: %d, expected: 1", list.Length)
	}
	if list.Head.Value != "a" {
		t.Fatalf("Head 값이 다릅니다. actual: %s, expected: a", list.Head.Value)
	}
	if list.Tail.Value != "a" {
		t.Fatalf("Tail 값이 다릅니다. actual: %s, expected: a", list.Tail.Value)
	}
}

func TestRPush_MultipleElements(t *testing.T) {
	// given
	list := NewList()

	// when: a, b, c 순서로 RPush → [a, b, c]
	list.RPush("a")
	list.RPush("b")
	list.RPush("c")

	// then
	if list.Length != 3 {
		t.Fatalf("길이가 다릅니다. actual: %d, expected: 3", list.Length)
	}
	if list.Head.Value != "a" {
		t.Fatalf("Head 값이 다릅니다. actual: %s, expected: a", list.Head.Value)
	}
	if list.Tail.Value != "c" {
		t.Fatalf("Tail 값이 다릅니다. actual: %s, expected: c", list.Tail.Value)
	}
}

func TestLPop_EmptyList(t *testing.T) {
	// given
	list := NewList()

	// when
	value, ok := list.LPop()

	// then
	if ok {
		t.Fatalf("빈 리스트에서 Pop 성공: %s", value)
	}
	if value != "" {
		t.Fatalf("빈 리스트에서 값이 반환됨: %s", value)
	}
}

func TestLPop_SingleElement(t *testing.T) {
	// given
	list := NewList()
	list.LPush("a")

	// when
	value, ok := list.LPop()

	// then
	if !ok {
		t.Fatal("Pop 실패")
	}
	if value != "a" {
		t.Fatalf("값이 다릅니다. actual: %s, expected: a", value)
	}
	if list.Length != 0 {
		t.Fatalf("길이가 다릅니다. actual: %d, expected: 0", list.Length)
	}
	if list.Head != nil {
		t.Fatal("Head가 nil이 아닙니다")
	}
	if list.Tail != nil {
		t.Fatal("Tail이 nil이 아닙니다")
	}
}

func TestLPop_MultipleElements(t *testing.T) {
	// given: [c, b, a]
	list := NewList()
	list.LPush("a")
	list.LPush("b")
	list.LPush("c")

	// when
	value, _ := list.LPop()

	// then: Head에서 제거 → c
	if value != "c" {
		t.Fatalf("값이 다릅니다. actual: %s, expected: c", value)
	}
	if list.Length != 2 {
		t.Fatalf("길이가 다릅니다. actual: %d, expected: 2", list.Length)
	}
	if list.Head.Value != "b" {
		t.Fatalf("Head 값이 다릅니다. actual: %s, expected: b", list.Head.Value)
	}
	if list.Head.Prev != nil {
		t.Fatal("새 Head의 Prev가 nil이 아닙니다")
	}
}

func TestRPop_EmptyList(t *testing.T) {
	// given
	list := NewList()

	// when
	value, ok := list.RPop()

	// then
	if ok {
		t.Fatalf("빈 리스트에서 Pop 성공: %s", value)
	}
}

func TestRPop_SingleElement(t *testing.T) {
	// given
	list := NewList()
	list.RPush("a")

	// when
	value, ok := list.RPop()

	// then
	if !ok {
		t.Fatal("Pop 실패")
	}
	if value != "a" {
		t.Fatalf("값이 다릅니다. actual: %s, expected: a", value)
	}
	if list.Length != 0 {
		t.Fatalf("길이가 다릅니다. actual: %d, expected: 0", list.Length)
	}
	if list.Head != nil {
		t.Fatal("Head가 nil이 아닙니다")
	}
	if list.Tail != nil {
		t.Fatal("Tail이 nil이 아닙니다")
	}
}

func TestRPop_MultipleElements(t *testing.T) {
	// given: [a, b, c]
	list := NewList()
	list.RPush("a")
	list.RPush("b")
	list.RPush("c")

	// when
	value, _ := list.RPop()

	// then: Tail에서 제거 → c
	if value != "c" {
		t.Fatalf("값이 다릅니다. actual: %s, expected: c", value)
	}
	if list.Length != 2 {
		t.Fatalf("길이가 다릅니다. actual: %d, expected: 2", list.Length)
	}
	if list.Tail.Value != "b" {
		t.Fatalf("Tail 값이 다릅니다. actual: %s, expected: b", list.Tail.Value)
	}
	if list.Tail.Next != nil {
		t.Fatal("새 Tail의 Next가 nil이 아닙니다")
	}
}

func TestLPushAndRPop(t *testing.T) {
	// given: LPush로 넣고 RPop으로 빼기 (큐처럼 동작)
	list := NewList()
	list.LPush("a")
	list.LPush("b")
	list.LPush("c")
	// 리스트: [c, b, a]

	// when & then: RPop은 Tail부터 → a, b, c 순서
	v1, _ := list.RPop()
	v2, _ := list.RPop()
	v3, _ := list.RPop()

	if v1 != "a" || v2 != "b" || v3 != "c" {
		t.Fatalf("순서가 다릅니다. actual: %s, %s, %s, expected: a, b, c", v1, v2, v3)
	}
	if list.Length != 0 {
		t.Fatalf("길이가 다릅니다. actual: %d, expected: 0", list.Length)
	}
}

func TestRange_FullRange(t *testing.T) {
	// given: [a, b, c]
	list := NewList()
	list.RPush("a")
	list.RPush("b")
	list.RPush("c")

	// when
	result := list.Range(0, -1)

	// then
	if len(result) != 3 {
		t.Fatalf("길이가 다릅니다. actual: %d, expected: 3", len(result))
	}
	if result[0] != "a" || result[1] != "b" || result[2] != "c" {
		t.Fatalf("값이 다릅니다. actual: %v", result)
	}
}

func TestRange_PartialRange(t *testing.T) {
	// given: [a, b, c, d, e]
	list := NewList()
	list.RPush("a")
	list.RPush("b")
	list.RPush("c")
	list.RPush("d")
	list.RPush("e")

	// when
	result := list.Range(1, 3)

	// then: [b, c, d]
	if len(result) != 3 {
		t.Fatalf("길이가 다릅니다. actual: %d, expected: 3", len(result))
	}
	if result[0] != "b" || result[1] != "c" || result[2] != "d" {
		t.Fatalf("값이 다릅니다. actual: %v", result)
	}
}

func TestRange_NegativeIndex(t *testing.T) {
	// given: [a, b, c, d, e]
	list := NewList()
	list.RPush("a")
	list.RPush("b")
	list.RPush("c")
	list.RPush("d")
	list.RPush("e")

	// when: -3 ~ -1 → 인덱스 2 ~ 4
	result := list.Range(-3, -1)

	// then: [c, d, e]
	if len(result) != 3 {
		t.Fatalf("길이가 다릅니다. actual: %d, expected: 3", len(result))
	}
	if result[0] != "c" || result[1] != "d" || result[2] != "e" {
		t.Fatalf("값이 다릅니다. actual: %v", result)
	}
}

func TestRange_OutOfBounds(t *testing.T) {
	// given: [a, b, c]
	list := NewList()
	list.RPush("a")
	list.RPush("b")
	list.RPush("c")

	// when: stop이 범위를 초과
	result := list.Range(0, 100)

	// then: 전체 반환
	if len(result) != 3 {
		t.Fatalf("길이가 다릅니다. actual: %d, expected: 3", len(result))
	}
}

func TestRange_StartGreaterThanStop(t *testing.T) {
	// given: [a, b, c]
	list := NewList()
	list.RPush("a")
	list.RPush("b")
	list.RPush("c")

	// when: start > stop
	result := list.Range(2, 0)

	// then: 빈 배열
	if len(result) != 0 {
		t.Fatalf("빈 배열이 아닙니다. actual: %v", result)
	}
}

func TestRange_EmptyList(t *testing.T) {
	// given
	list := NewList()

	// when
	result := list.Range(0, -1)

	// then
	if len(result) != 0 {
		t.Fatalf("빈 배열이 아닙니다. actual: %v", result)
	}
}
