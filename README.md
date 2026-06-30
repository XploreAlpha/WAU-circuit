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

## v0.9.0 "Acorn" 收口段(2026-09-15 GA)

上文详细介绍了 circuit domain DSL 本体。本段为 v0.9.0 GA 增量补充。

### 角色

| OS 类比 | Domain Specific Layer(DSL 解释器)|
|---|---|
| 部署 | 独立 git 仓 = `wau-circuit`,WAU-core-kernel 6 子模块之一 |
| 通信 | 通过 WAU-core-kernel 间接调用,wasm interpreter |
| 状态 | v0.8.0 GA 已发(2026-07-13)|

### v0.9.0 集成

- **电路域 DSL**(per [[project-wau-network-society-vision-2026-06-29]]):用于描述 agent 之间的连接图
- **wasm interpreter** 沙箱执行,保证安全
- **v1.x 升级时**:可见 DSO 化(per vision 文档)

### v0.9.0 "Acorn" 5 份核心文档

| # | 文件 | 内容 |
|---|---|---|
| 1 | [README.md](README.md)(本文件)| 仓入口 + DSL 设计 + v0.9.0 收口段 |
| 2 | [QUICKSTART.md](QUICKSTART.md) | 15 分钟跑通 1 个 circuit 描述 |
| 3 | [DEPLOY.md](DEPLOY.md) | wasm 运行时 + 解释器配置 |
| 4 | [ARCHITECTURE.md](ARCHITECTURE.md) | DSL 语法 + IR + 解释器 |
| 5 | [CHANGELOG.md](CHANGELOG.md) | v0.8.0 + v0.9.0 倒序 |

## License

MIT License - 详见 [LICENSE](LICENSE) 文件
