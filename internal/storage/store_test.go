package storage

import (
	"fmt"
	"sync"
	"testing"
)

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

func TestConcurrentSetGet(t *testing.T) {
    // given
    store := New()
    var wg sync.WaitGroup

    // when: 여러 고루틴이 동시에 SET/GET 수행
    for i := 0; i < 100; i++ {
        wg.Add(2)

        go func(n int) {
            defer wg.Done()
            key := fmt.Sprintf("key-%d", n)
            store.Set(key, "value")
        }(i)

        go func(n int) {
            defer wg.Done()
            key := fmt.Sprintf("key-%d", n)
            store.Get(key)
        }(i)
    }

    wg.Wait()
    // then: -race 플래그로 실행 시 race가 감지되지 않아야 함
}

func TestConcurrentWrite(t *testing.T) {
    // given
    store := New()
    var wg sync.WaitGroup

    // when: 100개의 고루틴이 같은 키에 동시 쓰기
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            store.Set("counter", fmt.Sprintf("%d", n))
        }(i)
    }

    wg.Wait()

    // then: 값이 존재해야 함 (어떤 값이든)
    value, exist := store.Get("counter")
    if !exist {
        t.Fatal("counter 키가 존재하지 않음")
    }
    // value는 0~99 중 하나 (마지막에 쓴 고루틴의 값)
    t.Logf("최종 값: %s", value)
}
