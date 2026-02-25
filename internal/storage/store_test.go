package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
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

// ===== TTL 관련 테스트 =====

func TestExpireAndTTL(t *testing.T) {
	// given
	store := New()
	store.Set("key", "value")

	// when
	result := store.Expire("key", 10)

	// then
	if result != 1 {
		t.Fatalf("Expire 결과가 다릅니다. actual: %d, expected: 1", result)
	}
	ttl := store.TTL("key")
	if ttl <= 0 || ttl > 10 {
		t.Fatalf("TTL 범위 초과. actual: %d", ttl)
	}
}

func TestExpireNonExistentKey(t *testing.T) {
	// given
	store := New()

	// when
	result := store.Expire("unknown", 10)

	// then
	if result != 0 {
		t.Fatalf("존재하지 않는 키에 Expire: %d, expected: 0", result)
	}
}

func TestTTL_NoExpiry(t *testing.T) {
	// given: TTL 없이 SET
	store := New()
	store.Set("key", "value")

	// when
	ttl := store.TTL("key")

	// then
	if ttl != -1 {
		t.Fatalf("TTL이 다릅니다. actual: %d, expected: -1", ttl)
	}
}

func TestTTL_NonExistentKey(t *testing.T) {
	// given
	store := New()

	// when
	ttl := store.TTL("unknown")

	// then
	if ttl != -2 {
		t.Fatalf("TTL이 다릅니다. actual: %d, expected: -2", ttl)
	}
}

func TestLazyDeletion(t *testing.T) {
	// given: TTL 1초 설정
	store := New()
	store.Set("session", "abc")
	store.Expire("session", 1)

	// when: 2초 대기 후 Get
	time.Sleep(2 * time.Second)
	value, exist := store.Get("session")

	// then: 만료되어 존재하지 않음
	if exist {
		t.Fatalf("만료된 키가 존재합니다: %s", value)
	}
}

func TestLazyDeletion_List(t *testing.T) {
	// given: 리스트에 TTL 설정
	store := New()
	store.LPush("mylist", "a", "b")
	store.Expire("mylist", 1)

	// when: 2초 대기 후 LRange
	time.Sleep(2 * time.Second)
	result, err := store.LRange("mylist", 0, -1)

	// then: 만료되어 빈 배열
	if err != nil {
		t.Fatalf("에러 발생: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("만료된 리스트가 존재합니다: %v", result)
	}
}

func TestDel(t *testing.T) {
	// given
	store := New()
	store.Set("key", "value")

	// when
	result := store.Del("key")

	// then
	if result != 1 {
		t.Fatalf("Del 결과: %d, expected: 1", result)
	}
	_, exist := store.Get("key")
	if exist {
		t.Fatal("삭제된 키가 존재합니다")
	}
}

func TestDel_NonExistentKey(t *testing.T) {
	// given
	store := New()

	// when
	result := store.Del("unknown")

	// then
	if result != 0 {
		t.Fatalf("Del 결과: %d, expected: 0", result)
	}
}

func TestDel_List(t *testing.T) {
	// given
	store := New()
	store.LPush("mylist", "a", "b", "c")

	// when
	result := store.Del("mylist")

	// then
	if result != 1 {
		t.Fatalf("Del 결과: %d, expected: 1", result)
	}
	values, _ := store.LRange("mylist", 0, -1)
	if len(values) != 0 {
		t.Fatalf("삭제된 리스트가 존재합니다: %v", values)
	}
}

func TestPersist(t *testing.T) {
	// given: EXPIRE 설정
	store := New()
	store.Set("key", "value")
	store.Expire("key", 10)

	// when
	result := store.Persist("key")

	// then: TTL 제거됨
	if result != 1 {
		t.Fatalf("Persist 결과: %d, expected: 1", result)
	}
	ttl := store.TTL("key")
	if ttl != -1 {
		t.Fatalf("Persist 후 TTL: %d, expected: -1", ttl)
	}
}

func TestPersist_NoExpiry(t *testing.T) {
	// given: TTL 없는 키
	store := New()
	store.Set("key", "value")

	// when
	result := store.Persist("key")

	// then
	if result != 0 {
		t.Fatalf("Persist 결과: %d, expected: 0", result)
	}
}

func TestPersist_NonExistentKey(t *testing.T) {
	// given
	store := New()

	// when
	result := store.Persist("unknown")

	// then
	if result != 0 {
		t.Fatalf("Persist 결과: %d, expected: 0", result)
	}
}

func TestExpireOverwrite(t *testing.T) {
	// given: EXPIRE 10초 설정
	store := New()
	store.Set("key", "value")
	store.Expire("key", 10)

	// when: EXPIRE 100초로 재설정
	store.Expire("key", 100)

	// then: TTL이 10초보다 큰 값
	ttl := store.TTL("key")
	if ttl <= 10 {
		t.Fatalf("TTL이 갱신되지 않았습니다. actual: %d", ttl)
	}
}

func TestActiveDeletion(t *testing.T) {
	// given: 백그라운드 만료 시작, TTL 1초 설정
	store := New()
	store.StartExpiry()
	defer store.StopExpiry()

	store.Set("session", "abc")
	store.Expire("session", 1)

	// when: 3초 대기 (1초 TTL + 1초 ticker 간격 + 여유)
	time.Sleep(3 * time.Second)

	// then: Active Deletion으로 키가 삭제됨
	ttl := store.TTL("session")
	if ttl != -2 {
		t.Fatalf("Active Deletion 실패. TTL: %d, expected: -2", ttl)
	}
}

func TestConcurrentExpire(t *testing.T) {
	// given
	store := New()
	var wg sync.WaitGroup

	// when: 여러 고루틴이 동시에 EXPIRE/TTL/DEL 수행
	for i := 0; i < 100; i++ {
		wg.Add(3)

		go func(n int) {
			defer wg.Done()
			key := fmt.Sprintf("ttl-key-%d", n)
			store.Set(key, "value")
			store.Expire(key, 60)
		}(i)

		go func(n int) {
			defer wg.Done()
			key := fmt.Sprintf("ttl-key-%d", n)
			store.TTL(key)
		}(i)

		go func(n int) {
			defer wg.Done()
			key := fmt.Sprintf("ttl-key-%d", n)
			store.Del(key)
		}(i)
	}

	wg.Wait()
	// then: -race 플래그로 실행 시 race가 감지되지 않아야 함
}

func TestSave_EmptyStore(t *testing.T) {
	// given: 빈 Store
	store := New()
	path := filepath.Join(t.TempDir(), "empty.rdb")

	// when
	err := store.Save(path)

	// then: Header(7) + EOF(1) + Checksum(4) = 12바이트
	if err != nil {
		t.Fatalf("에러 발생: %v", err)
	}
	data, _ := os.ReadFile(path)
	if len(data) != 12 {
		t.Fatalf("파일 크기: %d, expected: 12", len(data))
	}
	if string(data[:6]) != "MINIDB" {
		t.Fatalf("Magic bytes: %s", string(data[:6]))
	}
	if data[7] != 0xFF {
		t.Fatalf("EOF: 0x%02x, expected: 0xFF", data[7])
	}
}

func TestSave_StringEntries(t *testing.T) {
	// given
	store := New()
	store.Set("key", "value")
	path := filepath.Join(t.TempDir(), "string.rdb")

	// when
	err := store.Save(path)

	// then: Header(7) + String("key"=3,"value"=5: 1+4+3+4+5+1=18) + EOF(1) + Checksum(4) = 30
	if err != nil {
		t.Fatalf("에러 발생: %v", err)
	}
	data, _ := os.ReadFile(path)
	if len(data) != 30 {
		t.Fatalf("파일 크기: %d, expected: 30", len(data))
	}
	// 엔트리 타입이 TypeString(0x00)
	if data[7] != 0x00 {
		t.Fatalf("Entry type: 0x%02x, expected: 0x00", data[7])
	}
}

func TestSave_ListEntries(t *testing.T) {
	// given
	store := New()
	store.RPush("mylist", "a", "b", "c")
	path := filepath.Join(t.TempDir(), "list.rdb")

	// when
	err := store.Save(path)

	// then: Header(7) + List("mylist"=6,["a","b","c"]: 1+4+6+4+(4+1)*3+1=31) + EOF(1) + Checksum(4) = 43
	if err != nil {
		t.Fatalf("에러 발생: %v", err)
	}
	data, _ := os.ReadFile(path)
	if len(data) != 43 {
		t.Fatalf("파일 크기: %d, expected: 43", len(data))
	}
	// 엔트리 타입이 TypeList(0x01)
	if data[7] != 0x01 {
		t.Fatalf("Entry type: 0x%02x, expected: 0x01", data[7])
	}
}

func TestSave_SkipExpiredKeys(t *testing.T) {
	// given: 만료된 키 1개 + 유효한 키 1개
	store := New()
	store.Set("expired", "old")
	store.Set("valid", "new")
	store.Expire("expired", 1)
	time.Sleep(2 * time.Second)

	path := filepath.Join(t.TempDir(), "skip.rdb")

	// when
	err := store.Save(path)

	// then: 만료 키 제외, "valid" 키만 저장
	// Header(7) + String("valid"=5,"new"=3: 1+4+5+4+3+1=18) + EOF(1) + Checksum(4) = 30
	if err != nil {
		t.Fatalf("에러 발생: %v", err)
	}
	data, _ := os.ReadFile(path)
	if len(data) != 30 {
		t.Fatalf("파일 크기: %d, expected: 30 (만료 키 제외)", len(data))
	}
}

func TestSave_WithTTL(t *testing.T) {
	// given: TTL이 설정된 키
	store := New()
	store.Set("session", "abc")
	store.Expire("session", 3600)
	path := filepath.Join(t.TempDir(), "ttl.rdb")

	// when
	err := store.Save(path)

	// then: Header(7) + String("session"=7,"abc"=3,+TTL: 1+4+7+4+3+1+8=28) + EOF(1) + Checksum(4) = 40
	if err != nil {
		t.Fatalf("에러 발생: %v", err)
	}
	data, _ := os.ReadFile(path)
	if len(data) != 40 {
		t.Fatalf("파일 크기: %d, expected: 40", len(data))
	}
	// HasExpiry 마커 확인: type(1)+keylen(4)+key(7)+vallen(4)+val(3) = offset 7+19 = 26
	if data[26] != 0x01 {
		t.Fatalf("HasExpiry: 0x%02x, expected: 0x01", data[26])
	}
}

// ========== Load 테스트 ==========

func TestLoad_NonExistentFile(t *testing.T) {
	// given: 존재하지 않는 파일 경로
	store := New()
	path := filepath.Join(t.TempDir(), "nofile.rdb")

	// when
	err := store.Load(path)

	// then: 에러 없이 정상 반환
	if err != nil {
		t.Fatalf("존재하지 않는 파일에 에러 발생: %v", err)
	}
}

func TestLoad_StringEntries(t *testing.T) {
	// given: Save로 String 엔트리 저장
	store := New()
	store.Set("name", "gopher")
	store.Set("lang", "go")
	path := filepath.Join(t.TempDir(), "string.rdb")
	store.Save(path)

	// when: 새 Store에 Load
	loaded := New()
	err := loaded.Load(path)

	// then
	if err != nil {
		t.Fatalf("에러 발생: %v", err)
	}
	v1, ok1 := loaded.Get("name")
	if !ok1 || v1 != "gopher" {
		t.Fatalf("name: %s, exist: %v", v1, ok1)
	}
	v2, ok2 := loaded.Get("lang")
	if !ok2 || v2 != "go" {
		t.Fatalf("lang: %s, exist: %v", v2, ok2)
	}
}

func TestLoad_ListEntries(t *testing.T) {
	// given: Save로 List 엔트리 저장
	store := New()
	store.RPush("fruits", "apple", "banana", "cherry")
	path := filepath.Join(t.TempDir(), "list.rdb")
	store.Save(path)

	// when: 새 Store에 Load
	loaded := New()
	err := loaded.Load(path)

	// then
	if err != nil {
		t.Fatalf("에러 발생: %v", err)
	}
	result, lErr := loaded.LRange("fruits", 0, -1)
	if lErr != nil {
		t.Fatalf("LRange 에러: %v", lErr)
	}
	expected := []string{"apple", "banana", "cherry"}
	if len(result) != len(expected) {
		t.Fatalf("리스트 길이: %d, expected: %d", len(result), len(expected))
	}
	for i, v := range result {
		if v != expected[i] {
			t.Fatalf("result[%d]: %s, expected: %s", i, v, expected[i])
		}
	}
}

func TestLoad_WithTTL(t *testing.T) {
	// given: TTL이 설정된 키를 Save
	store := New()
	store.Set("session", "abc")
	store.Expire("session", 3600) // 1시간 후 만료
	path := filepath.Join(t.TempDir(), "ttl.rdb")
	store.Save(path)

	// when: 새 Store에 Load
	loaded := New()
	err := loaded.Load(path)

	// then: 키가 존재하고 TTL이 복원됨
	if err != nil {
		t.Fatalf("에러 발생: %v", err)
	}
	v, ok := loaded.Get("session")
	if !ok || v != "abc" {
		t.Fatalf("session: %s, exist: %v", v, ok)
	}
	ttl := loaded.TTL("session")
	if ttl <= 0 {
		t.Fatalf("TTL: %d, 양수여야 합니다", ttl)
	}
}

func TestLoad_SkipExpiredKeys(t *testing.T) {
	// given: 만료된 키 + 유효한 키를 Save
	store := New()
	store.Set("expired", "old")
	store.Set("valid", "new")
	store.Expire("expired", 1)
	time.Sleep(2 * time.Second)

	path := filepath.Join(t.TempDir(), "expired.rdb")
	store.Save(path)

	// when: 새 Store에 Load
	loaded := New()
	err := loaded.Load(path)

	// then: 만료된 키는 로드되지 않음
	if err != nil {
		t.Fatalf("에러 발생: %v", err)
	}
	_, ok := loaded.Get("expired")
	if ok {
		t.Fatal("만료된 키가 로드되면 안 됩니다")
	}
	v, ok := loaded.Get("valid")
	if !ok || v != "new" {
		t.Fatalf("valid: %s, exist: %v", v, ok)
	}
}

func TestLoad_CorruptedChecksum(t *testing.T) {
	// given: Save 후 파일 변조
	store := New()
	store.Set("key", "value")
	path := filepath.Join(t.TempDir(), "corrupted.rdb")
	store.Save(path)

	data, _ := os.ReadFile(path)
	data[10] ^= 0xFF
	os.WriteFile(path, data, 0644)

	// when: 변조된 파일 Load
	loaded := New()
	err := loaded.Load(path)

	// then: 체크섬 에러
	if err == nil {
		t.Fatal("변조된 파일에 대해 에러가 발생해야 합니다")
	}
}

func TestLoad_EmptyStore(t *testing.T) {
	// given: 빈 Store를 Save
	store := New()
	path := filepath.Join(t.TempDir(), "empty.rdb")
	store.Save(path)

	// when: 새 Store에 Load
	loaded := New()
	err := loaded.Load(path)

	// then: 에러 없이 빈 상태
	if err != nil {
		t.Fatalf("에러 발생: %v", err)
	}
	_, ok := loaded.Get("anything")
	if ok {
		t.Fatal("빈 Store에서 키가 존재하면 안 됩니다")
	}
}
