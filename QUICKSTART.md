# wau-circuit 15 分钟跑通

> 目标:本机启 wau-circuit + 写 1 个简单 circuit 描述 + 验证解释器能解析。

## 前置

- Go 1.21+
- 端口 :18470(gRPC,如启用)/ :18471(HTTP)
- 一个简单的 circuit YAML 描述

## 步骤

### 1. 拉源码

```bash
cd ~/project/wau-circuit
git pull origin main
make build
ls bin/
```

### 2. 写 1 个最小 circuit 描述

```bash
mkdir -p ~/.wau
cat > ~/.wau/circuit-simple.yaml <<EOF
name: hello-circuit
nodes:
  - id: input
    type: producer
  - id: transformer
    type: function
    fn: "echo"
  - id: sink
    type: consumer
edges:
  - from: input
    to: transformer
  - from: transformer
    to: sink
EOF
```

### 3. 启

```bash
./bin/wau-circuit -config ~/.wau/circuit.yaml
# 预期:[wau-circuit] interpreter ready
```

### 4. 通过 WAU-core-kernel 跑 1 个 circuit

在另一个终端,调 WAU-core-kernel 的 ExecuteCircuit RPC,传 `hello-circuit` 名字 + 入参。

预期:circuit 跑通,日志中看到 producer / transformer / consumer 顺序触发

## 下一步

- [DEPLOY.md](DEPLOY.md) — wasm 部署 + 解释器配置
- [ARCHITECTURE.md](ARCHITECTURE.md) — DSL 语法 + IR 设计
- [README.md](README.md) — v0.9.0 收口段
