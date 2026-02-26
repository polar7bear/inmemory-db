package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"time"
)

func main() {
	addr := flag.String("addr", "localhost:6379", "서버 주소")
	clients := flag.Int("c", 50, "동시 클라이언트 수")
	requests := flag.Int("n", 10000, "총 요청 수")
	tests := flag.String("t", "set,get", "테스트할 명령어 (쉼표 구분)")
	flag.Parse()

	commands := strings.Split(*tests, ",")
	for _, cmd := range commands {
		cmd = strings.TrimSpace(strings.ToLower(cmd))
		latencies, elapsed := runBenchmark(*addr, *clients, *requests, cmd)
		printResult(cmd, latencies, elapsed, *clients)
	}
}

// RESP 프로토콜 형식의 명령어 바이트를 생성한다.
// buildCommand("SET", "key:000001", "value")
// -> "*3\r\n$3\r\nSET\r\n$10\r\nkey:000001\r\n$5\r\nvalue\r\n"
func buildCommand(args ...string) []byte {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("*%d\r\n", len(args)))
	for _, arg := range args {
		buf.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg))
	}
	return buf.Bytes()
}

// 명령어 종류에 따라 요청 번호(i)를 받아 RESP 바이트를 반환한다.
func buildRequest(cmd string, i int) []byte {
	key := fmt.Sprintf("key:%06d", i)
	switch cmd {
	case "set":
		return buildCommand("SET", key, "value")
	case "get":
		return buildCommand("GET", key)
	case "lpush":
		return buildCommand("LPUSH", "bench-list", fmt.Sprintf("val:%d", i))
	case "rpush":
		return buildCommand("RPUSH", "bench-list", fmt.Sprintf("val:%d", i))
	default:
		return buildCommand("PING")
	}
}

// 응답 타입에 따라 적절한 줄 수만큼 소비한다.
// Simple String(+), Error(-), Integer(:) -> 1줄 (이미 읽음)
// Bulk String($) -> 추가 1줄 읽기 ($-1은 제외)
func consumeResponse(reader *bufio.Reader) {
	line, _ := reader.ReadString('\n')
	// Bulk String이면 데이터 줄을 한 번 더 읽어야 한다
	if len(line) > 0 && line[0] == '$' && line[1] != '-' {
		reader.ReadString('\n')
	}
}

// worker는 하나의 TCP 연결에서 할당된 요청을 순차 실행하고
// 각 요청의 latency를 채널로 보낸다.
func worker(addr string, cmd string, startIdx, count int, results chan<- time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Printf("연결 실패: %v\n", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for i := startIdx; i < startIdx+count; i++ {
		req := buildRequest(cmd, i)

		start := time.Now()
		conn.Write(req)
		consumeResponse(reader)
		results <- time.Since(start)
	}
}

// 전체 벤치마크를 실행한다.
// clients개의 worker 고루틴이 requests개의 요청을 나눠서 처리한다.
func runBenchmark(addr string, clients, requests int, cmd string) ([]time.Duration, time.Duration) {
	// GET 벤치마크 전에 키를 미리 생성한다
	if cmd == "get" {
		seedKeys(addr, requests)
	}

	results := make(chan time.Duration, requests)
	var wg sync.WaitGroup

	perClient := requests / clients
	remainder := requests % clients

	wallStart := time.Now()

	for c := 0; c < clients; c++ {
		count := perClient
		if c < remainder {
			count++
		}
		startIdx := c * perClient
		if c < remainder {
			startIdx += c
		} else {
			startIdx += remainder
		}

		wg.Add(1)
		go worker(addr, cmd, startIdx, count, results, &wg)
	}

	// 모든 worker 완료 후 채널을 닫는다
	go func() {
		wg.Wait()
		close(results)
	}()

	latencies := make([]time.Duration, 0, requests)
	for d := range results {
		latencies = append(latencies, d)
	}

	elapsed := time.Since(wallStart)
	return latencies, elapsed
}

// GET 벤치마크를 위해 키를 미리 SET한다.
func seedKeys(addr string, count int) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for i := 0; i < count; i++ {
		conn.Write(buildRequest("set", i))
		consumeResponse(reader)
	}
}

// 정렬된 latency 슬라이스에서 p번째 백분위수를 반환한다.
func percentile(sorted []time.Duration, p float64) time.Duration {
	index := int(float64(len(sorted)) * p)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	return sorted[index]
}

// 벤치마크 결과를 출력한다.
func printResult(cmd string, latencies []time.Duration, elapsed time.Duration, clients int) {
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	total := len(latencies)
	throughput := float64(total) / elapsed.Seconds()

	fmt.Printf("\n====== %s ======\n", strings.ToUpper(cmd))
	fmt.Printf("  %d requests completed in %.2f seconds\n", total, elapsed.Seconds())
	fmt.Printf("  %d parallel clients\n", clients)
	fmt.Println()
	fmt.Println("  Latency:")
	fmt.Printf("    p50:  %.3f ms\n", float64(percentile(latencies, 0.50).Microseconds())/1000)
	fmt.Printf("    p95:  %.3f ms\n", float64(percentile(latencies, 0.95).Microseconds())/1000)
	fmt.Printf("    p99:  %.3f ms\n", float64(percentile(latencies, 0.99).Microseconds())/1000)
	fmt.Printf("    max:  %.3f ms\n", float64(latencies[total-1].Microseconds())/1000)
	fmt.Println()
	fmt.Printf("  Throughput: %.2f requests/sec\n", throughput)
	fmt.Println()
}
