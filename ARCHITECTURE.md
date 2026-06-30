# wau-circuit 架构

## 模块拆分

```
wau-circuit/
├── cmd/wau-circuit/main.go     # 主入口
├── internal/
│   ├── dsl/                    # DSL 解析(YAML)
│   ├── ir/                     # 中间表示
│   ├── interpreter/            # wasm 解释器(wazero)
│   └── metrics/                # prom 指标占位
├── examples/                    # 示例 circuit
├── tests/
└── README.md / QUICKSTART.md / DEPLOY.md / ARCHITECTURE.md / CHANGELOG.md
```

## 数据流(per [[project-wau-network-society-vision-2026-06-29]])

```
YAML 描述
    ↓ dsl.Parse
IR(nodes + edges)
    ↓ interpreter.Run(wasm bytecode)
节点 1 → 节点 2 → ... → 节点 N
    ↓ 状态机轮转结果
WAU-core-kernel 监听器
```

## 关键决策

| 决策 | 内容 |
|---|---|
| **6 子模块之一** | per [[project-wau-core-product-list-2026-06-28]] |
| **wasm wazero** | Go 原生,无 cgo 依赖,无系统 wasm runtime |
| **DSL YAML** | 不另起语法,用 YAML 描述 graph |

## 接口边界

- **入**:YAML circuit 描述文件
- **出**:执行结果(节点事件流 + 终态)
- **依赖**:无外部
- **被依赖**:WAU-core-kernel 转发 circuit 执行请求

## 性能预算

| 指标 | 目标 |
|---|---|
| DSL 解析 | < 10 ms |
| 单节点执行 | < 5 ms |
| circuit 触发延迟 | < 50 ms |

## 跟其他仓的关系

- **上游(调用本仓)**:WAU-core-kernel
- **下游**:无
- **同组**:wau-scheduler / wau-trust / wau-profile / wau-intent / wau-registry / wau-registry-service
