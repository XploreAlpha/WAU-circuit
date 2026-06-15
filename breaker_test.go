// Package circuit_test — table-driven 单测覆盖 5 状态转移
//
// 目的:M3 W5.2 任务"wau-circuit 补单测",为 3 语言 SDK 翻译做地基
// 覆盖:Closed → Open → HalfOpen → Closed / HalfOpen → Open / 变参 IsOpen / Reset
package circuit_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wau/circuit"
)

// ============================
// 状态机:Closed → Open
// ============================

func TestBreaker_ClosedToOpen_AfterThresholdFailures(t *testing.T) {
	cb := circuit.NewBreaker(nil)
	cb.SetFailureThreshold(3)
	cb.SetRecoveryTimeout(50 * time.Millisecond)

	// 2 次失败:仍 Closed
	for i := 0; i < 2; i++ {
		cb.RecordFailure("agent-A")
		if got := cb.GetState("agent-A"); got != circuit.CircuitClosed {
			t.Fatalf("after %d failures expected Closed, got %s", i+1, got)
		}
	}

	// 第 3 次:跳 Open
	cb.RecordFailure("agent-A")
	if got := cb.GetState("agent-A"); got != circuit.CircuitOpen {
		t.Fatalf("after 3 failures expected Open, got %s", got)
	}
}

// ============================
// 状态机:Open → HalfOpen(超时后)
// ============================

func TestBreaker_OpenToHalfOpen_AfterRecoveryTimeout(t *testing.T) {
	cb := circuit.NewBreaker(nil)
	cb.SetFailureThreshold(1)
	cb.SetRecoveryTimeout(20 * time.Millisecond)

	cb.RecordFailure("agent-B")
	if got := cb.GetState("agent-B"); got != circuit.CircuitOpen {
		t.Fatalf("expected Open after 1 failure, got %s", got)
	}

	// 立即查:仍 Open(GetState 内部检查 time.Since 不通过)
	if got := cb.GetState("agent-B"); got != circuit.CircuitOpen {
		t.Fatalf("expected Open immediately, got %s", got)
	}

	// 等超时
	time.Sleep(30 * time.Millisecond)

	// 现在 GetState 应触发 Open → HalfOpen
	if got := cb.GetState("agent-B"); got != circuit.CircuitHalfOpen {
		t.Fatalf("expected HalfOpen after timeout, got %s", got)
	}
}

// ============================
// 状态机:HalfOpen → Closed(成功)
// ============================

func TestBreaker_HalfOpenToClosed_OnSuccess(t *testing.T) {
	cb := circuit.NewBreaker(nil)
	cb.SetFailureThreshold(1)
	cb.SetRecoveryTimeout(10 * time.Millisecond)

	cb.RecordFailure("agent-C")
	time.Sleep(15 * time.Millisecond)
	_ = cb.GetState("agent-C") // 触发 Open → HalfOpen
	if got := cb.GetState("agent-C"); got != circuit.CircuitHalfOpen {
		t.Fatalf("expected HalfOpen, got %s", got)
	}

	cb.RecordSuccess("agent-C")
	if got := cb.GetState("agent-C"); got != circuit.CircuitClosed {
		t.Fatalf("expected Closed after success in HalfOpen, got %s", got)
	}
}

// ============================
// 状态机:HalfOpen → Open(再失败)
// ============================

func TestBreaker_HalfOpenToOpen_OnFailure(t *testing.T) {
	cb := circuit.NewBreaker(nil)
	cb.SetFailureThreshold(1)
	cb.SetRecoveryTimeout(10 * time.Millisecond)

	cb.RecordFailure("agent-D")
	time.Sleep(15 * time.Millisecond)
	_ = cb.GetState("agent-D") // 触发 Open → HalfOpen

	// HalfOpen 状态下再失败:回 Open
	cb.RecordFailure("agent-D")
	if got := cb.GetState("agent-D"); got != circuit.CircuitOpen {
		t.Fatalf("expected Open after failure in HalfOpen, got %s", got)
	}
}

// ============================
// 未知 agent 默 Closed
// ============================

func TestBreaker_UnknownAgent_DefaultsClosed(t *testing.T) {
	cb := circuit.NewBreaker(nil)
	if got := cb.GetState("agent-zzz"); got != circuit.CircuitClosed {
		t.Fatalf("expected Closed for unknown agent, got %s", got)
	}
}

// ============================
// IsOpen 变参:任一 Open 即 true
// ============================

func TestBreaker_IsOpen_Variadic(t *testing.T) {
	cb := circuit.NewBreaker(nil)
	cb.SetFailureThreshold(1)

	cb.RecordFailure("agent-A") // Open
	// agent-B 未触发:仍 Closed

	// 包含 Open:应 true
	if !cb.IsOpen("agent-A", "agent-B") {
		t.Fatal("IsOpen(A, B) should be true (A is Open)")
	}
	if !cb.IsOpen("agent-A") {
		t.Fatal("IsOpen(A) should be true")
	}
	if cb.IsOpen("agent-B", "agent-C") {
		t.Fatal("IsOpen(B, C) should be false (both Closed)")
	}
	if cb.IsOpen() {
		t.Fatal("IsOpen() with no args should be false")
	}
}

// ============================
// Reset 清理状态
// ============================

func TestBreaker_Reset_ClearsState(t *testing.T) {
	cb := circuit.NewBreaker(nil)
	cb.SetFailureThreshold(1)

	cb.RecordFailure("agent-A") // Open
	if got := cb.GetState("agent-A"); got != circuit.CircuitOpen {
		t.Fatalf("expected Open, got %s", got)
	}

	cb.Reset("agent-A")
	if got := cb.GetState("agent-A"); got != circuit.CircuitClosed {
		t.Fatalf("expected Closed after Reset, got %s", got)
	}
}

// ============================
// 并发安全(10 goroutine × 1000 Record*)
// ============================

func TestBreaker_Concurrent_RecordSafe(t *testing.T) {
	cb := circuit.NewBreaker(nil)
	cb.SetFailureThreshold(1000)

	var wg sync.WaitGroup
	var successCount, failCount int64

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				if j%2 == 0 {
					cb.RecordFailure("agent-concurrent")
					atomic.AddInt64(&failCount, 1)
				} else {
					cb.RecordSuccess("agent-concurrent")
					atomic.AddInt64(&successCount, 1)
				}
			}
		}()
	}
	wg.Wait()

	if successCount != 5000 || failCount != 5000 {
		t.Fatalf("expected 5000/5000, got %d/%d", successCount, failCount)
	}
	// 状态在 1000 阈值下应仍 Closed(交替 success/fail 互相抵消)
	if got := cb.GetState("agent-concurrent"); got != circuit.CircuitClosed {
		t.Fatalf("expected Closed after alternating, got %s", got)
	}
}

// ============================
// String 格式化
// ============================

func TestCircuitState_String(t *testing.T) {
	cases := []struct {
		state circuit.CircuitState
		want  string
	}{
		{circuit.CircuitClosed, "closed"},
		{circuit.CircuitOpen, "open"},
		{circuit.CircuitHalfOpen, "half-open"},
		{circuit.CircuitState(99), "closed"}, // 未知值兜底
	}
	for _, c := range cases {
		if got := c.state.String(); got != c.want {
			t.Errorf("CircuitState(%d).String() = %q, want %q", c.state, got, c.want)
		}
	}
}

// ============================
// GetAllStates — v0.7.0 W5.2.2 新增
// ============================

func TestBreaker_GetAllStates(t *testing.T) {
	t.Run("empty breaker returns empty map", func(t *testing.T) {
		cb := circuit.NewBreaker(nil)
		got := cb.GetAllStates()
		if got == nil {
			t.Fatal("expected non-nil empty map, got nil")
		}
		if len(got) != 0 {
			t.Errorf("expected empty, got %d entries", len(got))
		}
	})

	t.Run("returns agents that have entered Open or HalfOpen", func(t *testing.T) {
		cb := circuit.NewBreaker(nil)
		cb.SetFailureThreshold(2)
		// A: 触发 Open(states[A]=Open)
		cb.RecordFailure("A")
		cb.RecordFailure("A")
		// B: 1 次失败,未到阈值,states[B] 没 key
		cb.RecordFailure("B")
		// C: RecordSuccess 不写 states
		cb.RecordSuccess("C")
		// D: RecordFailure 2 次也 Open
		cb.RecordFailure("D")
		cb.RecordFailure("D")

		got := cb.GetAllStates()
		// 只 A、D 进了 states map(B 1 次失败 / C 0 次都没写)
		if len(got) != 2 {
			t.Fatalf("expected 2 entries (A, D), got %d: %+v", len(got), got)
		}
		if got["A"] != circuit.CircuitOpen {
			t.Errorf("A: want Open, got %s", got["A"])
		}
		if got["D"] != circuit.CircuitOpen {
			t.Errorf("D: want Open, got %s", got["D"])
		}
		if _, ok := got["B"]; ok {
			t.Errorf("B (1 failure, not at threshold) should NOT be in states map")
		}
		if _, ok := got["C"]; ok {
			t.Errorf("C (success only) should NOT be in states map")
		}
	})

	t.Run("returned map is a defensive copy", func(t *testing.T) {
		cb := circuit.NewBreaker(nil)
		cb.RecordFailure("X")
		cb.RecordSuccess("X")

		got := cb.GetAllStates()
		// 改返回 map 不应影响内部状态
		got["X"] = circuit.CircuitOpen
		got["Y"] = circuit.CircuitOpen // 新增不影响
		delete(got, "X")

		// 重新 GetAllStates,内部状态应不变
		again := cb.GetAllStates()
		if _, ok := again["Y"]; ok {
			t.Error("adding to returned map leaked into breaker state")
		}
		if again["X"] == circuit.CircuitOpen {
			t.Error("modifying returned map value leaked into breaker state")
		}
	})
}
