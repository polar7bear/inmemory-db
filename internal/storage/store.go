package storage


type Store struct {
	data map[string]string
}

func New() *Store {
	return &Store{data: make(map[string]string)}
}

func (s *Store) Set(key, value string) {
	s.data[key] = value
}

func (s *Store) Get(key string) (string, bool) {
	value, exist := s.data[key]
	return value, exist
}