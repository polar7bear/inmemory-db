package storage

import "time"

type HeapItem struct {
	Key      string
	ExpireAt time.Time
}

type MinHeap struct {
	items []*HeapItem
}

func NewMinHeap() *MinHeap {
	return &MinHeap{}
}

// 새 항목을 추가하고 Bubble Up으로 힙 속성을 복원한다. O(log n)
func (h *MinHeap) Push(item *HeapItem) {
	h.items = append(h.items, item)

	curIdx := len(h.items) - 1
	parentIdx := (curIdx - 1) / 2

	for curIdx != 0 && h.items[curIdx].ExpireAt.Before(h.items[parentIdx].ExpireAt) {
		h.items[parentIdx], h.items[curIdx] = h.items[curIdx], h.items[parentIdx]
		curIdx = parentIdx
		parentIdx = (curIdx - 1) / 2

	}
}

// 최솟값(가장 먼저 만료되는 항목)을 제거하고 반환한다.
// Bubble Down으로 힙 속성을 복원한다. O(log n)
// 힙이 비어있으면 nil을 반환한다.
func (h *MinHeap) Pop() *HeapItem {
	if len(h.items) == 0 {
		return nil
	}

	min := h.items[0]
	last := len(h.items) - 1

	h.items[0] = h.items[last]
	h.items = h.items[:last]

	curIdx := 0
	for {
		left := 2*curIdx + 1
		right := 2*curIdx + 2
		smallest := curIdx

		if left < len(h.items) && h.items[left].ExpireAt.Before(h.items[smallest].ExpireAt) {
			smallest = left
		}
		if right < len(h.items) && h.items[right].ExpireAt.Before(h.items[smallest].ExpireAt) {
			smallest = right
		}

		if smallest == curIdx {
			break
		}

		h.items[curIdx], h.items[smallest] = h.items[smallest], h.items[curIdx]
		curIdx = smallest
	}

	return min
}

// 최솟값을 제거하지 않고 조회한다. O(1)
// 힙이 비어있으면 nil을 반환한다.
func (h *MinHeap) Peek() *HeapItem {
	if len(h.items) == 0 {
		return nil
	}
	return h.items[0]
}

// 힙의 크기를 반환한다.
func (h *MinHeap) Len() int {
	return len(h.items)
}
