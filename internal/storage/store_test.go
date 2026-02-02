package storage

import "testing"

func TestSetAndGet(t *testing.T) {
	// given
	store := New()

	// when
	store.Set("name", "승기")

	// then
	value, exist := store.Get("name")

	if !exist {
		t.Fatalf("존재하지 않는 데이터: %s", value)
	}

	if value != "승기" {
		t.Fatalf("일치하지 않는 데이터: %s", value)
	}
}

func TestGetNonExistent(t *testing.T) {
	// given
	store := New()

	// when
	value, exist := store.Get("unknown")

	// then
	if value != "" {
		t.Fatalf("데이터가 존재합니다: %s", value)
	}

	if exist {
		t.Fatalf("데이터가 존재합니다")
	}
}

func TestOverWrite(t *testing.T) {
	// given
	store := New()
	store.Set("animal", "dog")

	// when
	store.Set("animal", "cat")

	// then
	value, _ := store.Get("animal")

	if value != "cat" {
		t.Fatalf("값 덮어쓰기 실패: %s", value)
	}
}
