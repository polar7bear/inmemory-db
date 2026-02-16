package storage

import (
	"errors"
	"sync"
	"time"
)

type EntryType int

var ErrWrongType = errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")

const (
	TypeString EntryType = iota
	TypeList
)

type Entry struct {
	Type     EntryType
	Str      string
	List     *List
	ExpireAt *time.Time
}
type Store struct {
	data map[string]*Entry
	mu   sync.RWMutex
	heap *MinHeap
	done chan struct{}
}

func New() *Store {
	return &Store{
		data: make(map[string]*Entry),
		heap: NewMinHeap(),
		done: make(chan struct{}),
	}
}

func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = &Entry{Type: TypeString, Str: value}
}

func (s *Store) Get(key string) (string, bool) {
	// 서로다른 고루틴 간의 읽기 작업에서는 블로킹 없이 동시에 통과
	// A 고루틴이 쓰기 작업 도중, B 고루틴이 데이터를 읽고 있다면 데이터 불일치 현상이 생길 수 있기때문에
	// RLock(읽기)이 걸려있으면 Lock(쓰기)은 대기, Lock이 걸려있으면 RLock은 대기
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exist := s.data[key]

	if !exist {
		return "", false
	} else {
		if entry.Type != TypeString {
			return "", false
		}
	}

	return entry.Str, exist
}

// 키가 존재하지 않으면 새 리스트를 생성한다
// 키가 존재하지만 TypeList가 아니면 ErrWrongType을 반환한다
func (s *Store) LPush(key string, values ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exist := s.data[key]

	if !exist {
		newEntry := Entry{Type: TypeList, List: NewList()}
		entry = &newEntry
		s.data[key] = entry
	}

	if entry.Type != TypeList {
		return 0, ErrWrongType
	}

	for _, v := range values {
		entry.List.LPush(v)
	}

	return entry.List.Length, nil
}

func (s *Store) RPush(key string, values ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exist := s.data[key]

	if !exist {
		newEntry := Entry{Type: TypeList, List: NewList()}
		entry = &newEntry
		s.data[key] = entry
	}

	if entry.Type != TypeList {
		return 0, ErrWrongType
	}

	for _, v := range values {
		entry.List.RPush(v)
	}

	return entry.List.Length, nil
}

// 빈 리스트가 되면 키를 삭제한다 (Redis 동작)
func (s *Store) LPop(key string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exist := s.data[key]

	if !exist {
		return "", exist, nil
	}

	if entry.Type != TypeList {
		return "", false, ErrWrongType
	}

	value, result := entry.List.LPop()

	if entry.List.Length == 0 {
		delete(s.data, key)
	}

	return value, result, nil
}
func (s *Store) RPop(key string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exist := s.data[key]

	if !exist {
		return "", exist, nil
	}

	if entry.Type != TypeList {
		return "", false, ErrWrongType
	}

	value, result := entry.List.RPop()

	if entry.List.Length == 0 {
		delete(s.data, key)
	}

	return value, result, nil
}

func (s *Store) LRange(key string, start, stop int) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exist := s.data[key]

	if !exist {
		return []string{}, nil
	}

	if entry.Type != TypeList {
		return nil, ErrWrongType
	}

	result := entry.List.Range(start, stop)

	return result, nil
}

// 키에 만료 시간을 설정한다.
// 키가 존재하면 1, 존재하지 않으면 0을 반환한다.
func (s *Store) Expire(key string, seconds int) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, exist := s.data[key]

	if exist {
		expire := time.Now().Add(time.Duration(seconds) * time.Second)
		entry.ExpireAt = &expire
		s.heap.Push(&HeapItem{Key: key, ExpireAt: expire})
		return 1
	} else {
		return 0
	}
}

// 키의 남은 수명(초)을 반환한다.
// TTL이 없으면 -1, 키가 존재하지 않으면 -2를 반환한다.
func (s *Store) TTL(key string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, exist := s.data[key]

	if exist {
		if entry.ExpireAt == nil {
			return -1
		}
		return int(time.Until(*entry.ExpireAt).Seconds())
	} else {
		return -2
	}
}

// 키를 삭제한다. 삭제된 키의 개수(0 또는 1)를 반환한다.
func (s *Store) Del(key string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exist := s.data[key]
	if !exist {
		return 0
	} else {
		delete(s.data, key)
		return 1
	}
}

// 키에서 만료 시간을 제거한다.
// TTL이 존재하고 제거했으면 1, 아니면 0을 반환한다.
func (s *Store) Persist(key string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, exist := s.data[key]

	if !exist || entry.ExpireAt == nil {
		return 0
	} else {
		entry.ExpireAt = nil
		return 1
	}
}
