package circuit

import (
	"log/slog"
	"sync"
	"time"
)

// CircuitState 熔断状态
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// String 返回状态字符串
func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "closed"
	}
}

const (
	// DefaultFailureThreshold 默认失败阈值
	DefaultFailureThreshold = 5

	// DefaultRecoveryTimeout 默认恢复超时
	DefaultRecoveryTimeout = 30 * time.Second
)

// Breaker 熔断器
type Breaker struct {
	logger *slog.Logger

	mu          sync.RWMutex
	states      map[string]CircuitState
	failures    map[string]int
	lastFailure map[string]time.Time

	failureThreshold int
	recoveryTimeout  time.Duration
}

// NewBreaker 创建熔断器
//
// logger 可为 nil — 此时自动 fallback 到 slog.Default(),避免 panic
// 这是 v0.6.0 M3 W5.2 修的 bug:SDK 翻译时 NewBreaker(nil) 是合法用法
func NewBreaker(logger *slog.Logger) *Breaker {
	if logger == nil {
		logger = slog.Default()
	}
	return &Breaker{
		logger:          logger,
		states:          make(map[string]CircuitState),
		failures:        make(map[string]int),
		lastFailure:     make(map[string]time.Time),
		failureThreshold: DefaultFailureThreshold,
		recoveryTimeout:  DefaultRecoveryTimeout,
	}
}

// RecordFailure 记录一次失败
func (cb *Breaker) RecordFailure(agentID string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures[agentID]++
	cb.lastFailure[agentID] = time.Now()

	// HalfOpen 状态再失败:直接回 Open(不需要等阈值)
	// 这是 v0.6.0 M3 W5.2 修的 bug:老代码只处理 Closed→Open,HalfOpen 失败后 state 不动,会继续放行流量
	if cb.states[agentID] == CircuitHalfOpen {
		cb.states[agentID] = CircuitOpen
		cb.logger.Warn("Circuit breaker re-opened from half-open",
			"agent", agentID,
		)
		return
	}

	if cb.states[agentID] == CircuitClosed && cb.failures[agentID] >= cb.failureThreshold {
		cb.states[agentID] = CircuitOpen
		cb.logger.Warn("Circuit breaker opened",
			"agent", agentID,
			"failures", cb.failures[agentID],
		)
	}
}

// RecordSuccess 记录一次成功
func (cb *Breaker) RecordSuccess(agentID string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures[agentID] = 0

	if cb.states[agentID] == CircuitHalfOpen {
		cb.states[agentID] = CircuitClosed
		cb.logger.Info("Circuit breaker closed",
			"agent", agentID,
		)
	}
}

// GetState 获取熔断状态
func (cb *Breaker) GetState(agentID string) CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	state, exists := cb.states[agentID]
	if !exists {
		return CircuitClosed
	}

	switch state {
	case CircuitClosed:
		return CircuitClosed
	case CircuitOpen:
		if lastFail, ok := cb.lastFailure[agentID]; ok {
			if time.Since(lastFail) > cb.recoveryTimeout {
				cb.states[agentID] = CircuitHalfOpen
				return CircuitHalfOpen
			}
		}
		return CircuitOpen
	case CircuitHalfOpen:
		return CircuitHalfOpen
	default:
		return CircuitClosed
	}
}

// IsOpen 检查是否熔断中
func (cb *Breaker) IsOpen(agentIDs ...string) bool {
	for _, id := range agentIDs {
		if cb.GetState(id) == CircuitOpen {
			return true
		}
	}
	return false
}

// Reset 重置熔断状态
func (cb *Breaker) Reset(agentID string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	delete(cb.states, agentID)
	delete(cb.failures, agentID)
	delete(cb.lastFailure, agentID)
}

// SetFailureThreshold 设置失败阈值
func (cb *Breaker) SetFailureThreshold(threshold int) {
	cb.failureThreshold = threshold
}

// SetRecoveryTimeout 设置恢复超时
func (cb *Breaker) SetRecoveryTimeout(timeout time.Duration) {
	cb.recoveryTimeout = timeout
}

// GetAllStates 返回所有 (agentID) 当前的熔断状态副本
//
// v0.7.0 W5.2.2 新增:补 v0.3.0 留下的 GetAllStates 占位(原 handler 返空 map)。
// 用途:运维 dashboard / 监控 / 跨 agent 健康检查。
//
// 行为:
//   - 返回 map 是内部 states 的 snapshot copy,外部修改不影响 Breaker 状态
//   - 没出现过的 agentID 不在结果里(只有 record 过失败/成功的才会出现)
//   - 读路径用 RLock,不阻塞 RecordXxx
//   - 已 stale 的 Open 状态(超过 recoveryTimeout)GetState 会转 HalfOpen,
//     这里直接返回内部存的 CircuitOpen,调用方需要的话再 GetState 单查
func (cb *Breaker) GetAllStates() map[string]CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	out := make(map[string]CircuitState, len(cb.states))
	for k, v := range cb.states {
		out[k] = v
	}
	return out
}
