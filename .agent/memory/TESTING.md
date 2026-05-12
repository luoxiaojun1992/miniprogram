# 测试与质量门禁

## 本地建议顺序
1. `go test -short ./...`
2. `go build -o /tmp/miniprogram-server ./cmd/server`
3. `golangci-lint run ./...`（可用时）

## Taskfile 命令
```bash
task dev
task build
task test
task test-unit
task test-coverage
task lint
task lint-fix
task ui-test
task ui-test-down
```

## CI 概览（`.github/workflows/ci.yml`）
1. Unit Tests（覆盖率阈值校验）
2. API Tests（k6）
3. UI Tests（Playwright，依赖 API Tests 成功后执行）
