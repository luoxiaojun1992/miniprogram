# 配置与启动

## 加载顺序
1. 读取环境变量 `CONFIG_PATH`
2. 若 `CONFIG_PATH` 非空，先加载配置文件
3. `APP_*` 环境变量始终覆盖配置项（Viper 规则）

## 关键配置项
- `server`：端口与模式
- `database`：MySQL 连接信息
- `redis`：Redis 连接信息
- `jwt`：密钥与过期时间
- `upload`：上传存储（local/cos）
- `rate_limit`：频控开关与阈值
- `debug.enable_test_token`：是否启用 `/v1/debug/token`（仅非生产）

## 配置样例
- 配置模板：`configs/config.yaml.sample`
