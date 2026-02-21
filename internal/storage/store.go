package storage

import (
	"errors"
	"inmemory-db/internal/persistence"
	"os"
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
	s.mu.Lock()
	defer s.mu.Unlock() // isExpired 내부함수에서 데이터 쓰기작업이 포함되어있어 Lock으로 변경

	entry, exist := s.data[key]

	if !exist {
		return "", false
	} else {
		if entry.Type != TypeString || s.isExpired(key) {
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

	if !exist || s.isExpired(key) {
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

	if !exist || s.isExpired(key) {
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

	if s.isExpired(key) {
		return "", false, nil
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

	if s.isExpired(key) {
		return "", false, nil
	}

	value, result := entry.List.RPop()

	if entry.List.Length == 0 {
		delete(s.data, key)
	}

	return value, result, nil
}

func (s *Store) LRange(key string, start, stop int) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exist := s.data[key]

	if !exist {
		return []string{}, nil
	}

	if entry.Type != TypeList {
		return nil, ErrWrongType
	}

	if s.isExpired(key) {
		return []string{}, nil
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

// 키가 만료되었는지 확인하고, 만료되었으면 삭제한다.
// mu.Lock()을 잡고 있는 상태에서 호출해야 한다 (내부용).
func (s *Store) isExpired(key string) bool {
	expire := s.data[key].ExpireAt

	if expire == nil {
		return false
	}

	if expire.Before(time.Now()) {
		delete(s.data, key)
		return true
	} else {
		return false
	}
}

// 백그라운드 만료 처리를 시작한다.
// 1초마다 힙을 확인하고 만료된 키를 삭제한다.
func (s *Store) StartExpiry() {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-s.done:
				return
			case <-ticker.C:
				s.mu.Lock()
				for {
					item := s.heap.Peek()
					if item == nil || !item.ExpireAt.Before(time.Now()) {
						break
					}
					s.heap.Pop()

					entry, exist := s.data[item.Key]
					if exist && entry.ExpireAt != nil && entry.ExpireAt.Equal(item.ExpireAt) {
						delete(s.data, item.Key)
					}
				}
				s.mu.Unlock()
			}
		}
	}()
}

// 백그라운드 만료 처리를 중지한다.
func (s *Store) StopExpiry() {
	close(s.done)
}

func (s *Store) Save(path string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := persistence.NewEncoder(file)
	encoder.WriteHeader()

	for key, entry := range s.data {
		// 이미 만료된 키는 저장할 필요 없으니 건너뛴다.
		// isExpired()는 삭제(쓰기)를 하기 때문에 RLock 상태에서 호출 불가
		// 읽기 전용으로 만료 여부만 확인한다.
		if entry.ExpireAt != nil && entry.ExpireAt.Before(time.Now()) {
			continue
		}

		switch entry.Type {
		case TypeString:
			encoder.WriteStringEntry(key, entry.Str, entry.ExpireAt)

		case TypeList:
			values := entry.List.Range(0, entry.List.Length-1)
			encoder.WriteListEntry(key, values, entry.ExpireAt)

		}
	}

	encoder.WriteEOF()
	return encoder.Flush()
}
