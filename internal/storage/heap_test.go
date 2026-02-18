package storage

import (
	"testing"
	"time"
)

func TestHeapPush(t *testing.T) {
	// given
	heap := NewMinHeap()
	now := time.Now()

	// when: 순서 없이 3개 추가
	heap.Push(&HeapItem{Key: "c", ExpireAt: now.Add(30 * time.Second)})
	heap.Push(&HeapItem{Key: "a", ExpireAt: now.Add(10 * time.Second)})
	heap.Push(&HeapItem{Key: "b", ExpireAt: now.Add(20 * time.Second)})

	// then: Peek은 가장 먼저 만료되는 "a"
	item := heap.Peek()
	if item.Key != "a" {
		t.Fatalf("루트가 다릅니다. actual: %s, expected: a", item.Key)
	}
	if heap.Len() != 3 {
		t.Fatalf("크기가 다릅니다. actual: %d, expected: 3", heap.Len())
	}
}

func TestHeapPop(t *testing.T) {
	// given
	heap := NewMinHeap()
	now := time.Now()
	heap.Push(&HeapItem{Key: "b", ExpireAt: now.Add(20 * time.Second)})
	heap.Push(&HeapItem{Key: "a", ExpireAt: now.Add(10 * time.Second)})
	heap.Push(&HeapItem{Key: "c", ExpireAt: now.Add(30 * time.Second)})

	// when: Pop
	item := heap.Pop()

	// then: 최솟값 "a" 제거, 다음 루트는 "b"
	if item.Key != "a" {
		t.Fatalf("Pop 결과가 다릅니다. actual: %s, expected: a", item.Key)
	}
	if heap.Len() != 2 {
		t.Fatalf("크기가 다릅니다. actual: %d, expected: 2", heap.Len())
	}
	next := heap.Peek()
	if next.Key != "b" {
		t.Fatalf("다음 루트가 다릅니다. actual: %s, expected: b", next.Key)
	}
}

func TestHeapPopOrder(t *testing.T) {
	// given: 5개 항목을 무작위 순서로 Push
	heap := NewMinHeap()
	now := time.Now()
	heap.Push(&HeapItem{Key: "d", ExpireAt: now.Add(40 * time.Second)})
	heap.Push(&HeapItem{Key: "a", ExpireAt: now.Add(10 * time.Second)})
	heap.Push(&HeapItem{Key: "e", ExpireAt: now.Add(50 * time.Second)})
	heap.Push(&HeapItem{Key: "b", ExpireAt: now.Add(20 * time.Second)})
	heap.Push(&HeapItem{Key: "c", ExpireAt: now.Add(30 * time.Second)})

	// when & then: Pop 순서가 만료 시간 순이어야 한다
	expected := []string{"a", "b", "c", "d", "e"}
	for i, key := range expected {
		item := heap.Pop()
		if item.Key != key {
			t.Fatalf("Pop 순서 오류 [%d]: actual: %s, expected: %s", i, item.Key, key)
		}
	}
}

func TestHeapPopEmpty(t *testing.T) {
	// given
	heap := NewMinHeap()

	// when
	item := heap.Pop()

	// then
	if item != nil {
		t.Fatalf("빈 힙에서 Pop이 nil이 아닙니다: %v", item)
	}
}

func TestHeapPeekEmpty(t *testing.T) {
	// given
	heap := NewMinHeap()

	// when
	item := heap.Peek()

	// then
	if item != nil {
		t.Fatalf("빈 힙에서 Peek이 nil이 아닙니다: %v", item)
	}
}

func TestHeapLen(t *testing.T) {
	// given
	heap := NewMinHeap()
	now := time.Now()

	// when & then: Push하면 증가, Pop하면 감소
	if heap.Len() != 0 {
		t.Fatalf("초기 크기: %d", heap.Len())
	}

	heap.Push(&HeapItem{Key: "a", ExpireAt: now.Add(10 * time.Second)})
	heap.Push(&HeapItem{Key: "b", ExpireAt: now.Add(20 * time.Second)})
	if heap.Len() != 2 {
		t.Fatalf("Push 후 크기: %d, expected: 2", heap.Len())
	}

	heap.Pop()
	if heap.Len() != 1 {
		t.Fatalf("Pop 후 크기: %d, expected: 1", heap.Len())
	}
}

func TestHeapBubbleUp(t *testing.T) {
	// given: 큰 값을 먼저 넣고
	heap := NewMinHeap()
	now := time.Now()
	heap.Push(&HeapItem{Key: "c", ExpireAt: now.Add(30 * time.Second)})
	heap.Push(&HeapItem{Key: "b", ExpireAt: now.Add(20 * time.Second)})

	// when: 가장 작은 값을 마지막에 Push
	heap.Push(&HeapItem{Key: "a", ExpireAt: now.Add(10 * time.Second)})

	// then: 루트로 올라와야 한다
	item := heap.Peek()
	if item.Key != "a" {
		t.Fatalf("Bubble Up 실패. 루트: %s, expected: a", item.Key)
	}
}

func TestHeapBubbleDown(t *testing.T) {
	// given
	heap := NewMinHeap()
	now := time.Now()
	heap.Push(&HeapItem{Key: "a", ExpireAt: now.Add(10 * time.Second)})
	heap.Push(&HeapItem{Key: "b", ExpireAt: now.Add(20 * time.Second)})
	heap.Push(&HeapItem{Key: "c", ExpireAt: now.Add(30 * time.Second)})

	// when: 루트 Pop
	heap.Pop()

	// then: "b"가 루트, "c"가 자식 — 힙 속성 유지
	item := heap.Peek()
	if item.Key != "b" {
		t.Fatalf("Bubble Down 실패. 루트: %s, expected: b", item.Key)
	}
}

func TestHeapSingleElement(t *testing.T) {
	// given
	heap := NewMinHeap()
	now := time.Now()
	heap.Push(&HeapItem{Key: "only", ExpireAt: now.Add(10 * time.Second)})

	// when
	item := heap.Pop()

	// then
	if item.Key != "only" {
		t.Fatalf("값이 다릅니다. actual: %s, expected: only", item.Key)
	}
	if heap.Len() != 0 {
		t.Fatalf("Pop 후 크기: %d, expected: 0", heap.Len())
	}
}
