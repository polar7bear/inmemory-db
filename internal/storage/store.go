package storage

import "sync"

type Store struct {
	data map[string]string
	mu   sync.RWMutex
}

func New() *Store {
	return &Store{data: make(map[string]string)}
}

func (s *Store) Set(key, value string) {
	s.mu.Lock()
	s.data[key] = value
	s.mu.Unlock()
}

func (s *Store) Get(key string) (string, bool) {
	// 서로다른 고루틴 간의 읽기 작업에서는 블로킹 없이 동시에 통과
	// A 고루틴이 쓰기 작업 도중, B 고루틴이 데이터를 읽고 있다면 데이터 불일치 현상이 생길 수 있기때문에
	// RLock(읽기)이 걸려있으면 Lock(쓰기)은 대기, Lock이 걸려있으면 RLock은 대기
	s.mu.RLock()
	value, exist := s.data[key]
	s.mu.RUnlock()

	return value, exist
}
