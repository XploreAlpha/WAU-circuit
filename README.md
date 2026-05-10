# wau-circuit

> WAU 网络的熔断器模块 - 防止失控 Agent 拖垮系统

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)](LICENSE)

---

## 核心设计

**熔断器状态机** - 自动防止故障扩散。

```
Closed (正常) ──→ Open (熔断) ──→ HalfOpen (探测) ──→ Closed (正常)
     ↑              │                │
     └──────────────┴────────────────┘
         探测成功后恢复
```

---

## 状态说明

| 状态 | 说明 | 触发条件 |
|------|------|----------|
| **Closed** | 正常，流量通过 | 默认状态 |
| **Open** | 熔断，拒绝流量 | 连续 5 次失败 |
| **HalfOpen** | 探测恢复 | Open 后 30 秒 |

---

## 接口设计

```go
type Breaker struct{}

// RecordFailure 记录失败
func (cb *Breaker) RecordFailure(agentID string)

// RecordSuccess 记录成功
func (cb *Breaker) RecordSuccess(agentID string)

// GetState 获取状态
func (cb *Breaker) GetState(agentID string) CircuitState

// IsOpen 检查是否熔断
func (cb *Breaker) IsOpen(agentIDs ...string) bool

// Reset 重置
func (cb *Breaker) Reset(agentID string)
```

---

## 使用示例

```go
breaker := circuit.NewBreaker(logger)

// Agent 失败
breaker.RecordFailure("benny")

// Agent 成功
breaker.RecordSuccess("benny")

// 检查是否熔断
if breaker.IsOpen("benny") {
    // 绕过 Benny，选择其他 Agent
}
```

---

## 项目结构

```
wau-circuit/
├── breaker.go    # 熔断器实现
├── go.mod
└── README.md
```

---

## License

MIT License - 详见 [LICENSE](LICENSE) 文件
