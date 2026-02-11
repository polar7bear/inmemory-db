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

func TestLPush_Basic(t *testing.T) {
	// given
	store := New()

	// when
	length, err := store.LPush("mylist", "a", "b", "c")

	// then
	if err != nil {
		t.Fatalf("에러 발생: %v", err)
	}
	if length != 3 {
		t.Fatalf("길이가 다릅니다. actual: %d, expected: 3", length)
	}
}

func TestRPush_Basic(t *testing.T) {
	// given
	store := New()

	// when
	length, err := store.RPush("mylist", "a", "b", "c")

	// then
	if err != nil {
		t.Fatalf("에러 발생: %v", err)
	}
	if length != 3 {
		t.Fatalf("길이가 다릅니다. actual: %d, expected: 3", length)
	}
}

func TestLPush_WrongType(t *testing.T) {
	// given: String 타입으로 키 설정
	store := New()
	store.Set("key", "value")

	// when: 같은 키에 LPush 시도
	_, err := store.LPush("key", "a")

	// then: WRONGTYPE 에러
	if err != ErrWrongType {
		t.Fatalf("에러가 다릅니다. actual: %v, expected: %v", err, ErrWrongType)
	}
}

func TestRPush_WrongType(t *testing.T) {
	// given
	store := New()
	store.Set("key", "value")

	// when
	_, err := store.RPush("key", "a")

	// then
	if err != ErrWrongType {
		t.Fatalf("에러가 다릅니다. actual: %v, expected: %v", err, ErrWrongType)
	}
}

func TestGet_WrongType(t *testing.T) {
	// given: List 타입으로 키 설정
	store := New()
	store.LPush("key", "a")

	// when: 같은 키에 GET 시도
	value, exist := store.Get("key")

	// then: 빈 문자열, false
	if exist {
		t.Fatalf("List 타입 키에 대해 exist가 true: %s", value)
	}
	if value != "" {
		t.Fatalf("값이 비어있지 않습니다: %s", value)
	}
}

func TestLPop_Basic(t *testing.T) {
	// given: LPush a, b, c → [c, b, a]
	store := New()
	store.LPush("mylist", "a", "b", "c")

	// when
	value, ok, err := store.LPop("mylist")

	// then: Head에서 제거 → c
	if err != nil {
		t.Fatalf("에러 발생: %v", err)
	}
	if !ok {
		t.Fatal("Pop 실패")
	}
	if value != "c" {
		t.Fatalf("값이 다릅니다. actual: %s, expected: c", value)
	}
}

func TestRPop_Basic(t *testing.T) {
	// given: LPush a, b, c → [c, b, a]
	store := New()
	store.LPush("mylist", "a", "b", "c")

	// when
	value, ok, err := store.RPop("mylist")

	// then: Tail에서 제거 → a
	if err != nil {
		t.Fatalf("에러 발생: %v", err)
	}
	if !ok {
		t.Fatal("Pop 실패")
	}
	if value != "a" {
		t.Fatalf("값이 다릅니다. actual: %s, expected: a", value)
	}
}

func TestLPop_NonExistentKey(t *testing.T) {
	// given
	store := New()

	// when
	value, ok, err := store.LPop("unknown")

	// then: 에러 아님, 값 없음
	if err != nil {
		t.Fatalf("에러가 발생했습니다: %v", err)
	}
	if ok {
		t.Fatalf("존재하지 않는 키에서 Pop 성공: %s", value)
	}
}

func TestRPop_NonExistentKey(t *testing.T) {
	// given
	store := New()

	// when
	value, ok, err := store.RPop("unknown")

	// then
	if err != nil {
		t.Fatalf("에러가 발생했습니다: %v", err)
	}
	if ok {
		t.Fatalf("존재하지 않는 키에서 Pop 성공: %s", value)
	}
}

func TestLPop_WrongType(t *testing.T) {
	// given
	store := New()
	store.Set("key", "value")

	// when
	_, _, err := store.LPop("key")

	// then
	if err != ErrWrongType {
		t.Fatalf("에러가 다릅니다. actual: %v, expected: %v", err, ErrWrongType)
	}
}

func TestRPop_WrongType(t *testing.T) {
	// given
	store := New()
	store.Set("key", "value")

	// when
	_, _, err := store.RPop("key")

	// then
	if err != ErrWrongType {
		t.Fatalf("에러가 다릅니다. actual: %v, expected: %v", err, ErrWrongType)
	}
}

func TestLPop_DeletesEmptyList(t *testing.T) {
	// given: 1개짜리 리스트
	store := New()
	store.LPush("mylist", "a")

	// when: Pop으로 비우기
	store.LPop("mylist")

	// then: 키 자체가 삭제됨 (다시 Pop하면 값 없음)
	_, ok, _ := store.LPop("mylist")
	if ok {
		t.Fatal("빈 리스트의 키가 삭제되지 않았습니다")
	}
}

func TestLRange_Basic(t *testing.T) {
	// given: LPush a, b, c → [c, b, a]
	store := New()
	store.LPush("mylist", "a", "b", "c")

	// when
	result, err := store.LRange("mylist", 0, -1)

	// then
	if err != nil {
		t.Fatalf("에러 발생: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("길이가 다릅니다. actual: %d, expected: 3", len(result))
	}
	if result[0] != "c" || result[1] != "b" || result[2] != "a" {
		t.Fatalf("값이 다릅니다. actual: %v", result)
	}
}

func TestLRange_NonExistentKey(t *testing.T) {
	// given
	store := New()

	// when
	result, err := store.LRange("unknown", 0, -1)

	// then
	if err != nil {
		t.Fatalf("에러가 발생했습니다: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("빈 배열이 아닙니다. actual: %v", result)
	}
}

func TestLRange_WrongType(t *testing.T) {
	// given
	store := New()
	store.Set("key", "value")

	// when
	_, err := store.LRange("key", 0, -1)

	// then
	if err != ErrWrongType {
		t.Fatalf("에러가 다릅니다. actual: %v, expected: %v", err, ErrWrongType)
	}
}

func TestConcurrentListOps(t *testing.T) {
	// given
	store := New()
	var wg sync.WaitGroup

	// when: 여러 고루틴이 동시에 List 연산 수행
	for i := 0; i < 100; i++ {
		wg.Add(2)

		go func(n int) {
			defer wg.Done()
			key := fmt.Sprintf("list-%d", n)
			store.LPush(key, "value")
		}(i)

		go func(n int) {
			defer wg.Done()
			key := fmt.Sprintf("list-%d", n)
			store.LRange(key, 0, -1)
		}(i)
	}

	wg.Wait()
	// then: -race 플래그로 실행 시 race가 감지되지 않아야 함
}
