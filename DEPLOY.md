# wau-circuit 部署

## 端口

| 端口 | 类型 | 端点 |
|---|---|---|
| 18470 | gRPC | `wau.circuit.v1.Interpreter`(如有)|
| 18471 | HTTP | `/healthz` + `/metrics` |

## wasm 运行时

wau-circuit 内嵌 wazero(wasm 解释器,Go 原生,无需 cgo)。

```bash
# 配置 wasm 内存上限(默认 256 MB)
WASM_MAX_MEMORY_MB=512

# 配置 circuit 执行超时(默认 30s)
CIRCUIT_EXEC_TIMEOUT_MS=30000
```

## 监控

```bash
curl -s http://localhost:18471/metrics | grep wau_circuit
```

## 进程管理

```bash
tmux new -d -s wau-circuit '/tmp/wau-circuit -config ~/.wau/circuit.yaml'
```

## 配置

| 字段 | 默认 | 说明 |
|---|---|---|
| `wasm.max_memory_mb` | `256` | wasm 内存上限 |
| `exec.timeout_ms` | `30000` | 单 circuit 执行超时 |
| `circuit.dir` | `~/.wau/circuits/` | circuit 定义文件目录 |

## 升级路径

- v0.9.0(Acorn)→ v0.8.0(Sprout):
  - circuit YAML schema 100% 兼容
- v0.9.0 → v1.0.0:DSO 化(per [[project-wau-network-society-vision-2026-06-29]])
