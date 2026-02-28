# Dujiao-Next API

Dujiao-Next API 是 Dujiao-Next 生态系统的后端 API 服务。它为电子商务平台提供完整的公共 API、用户认证、订单管理、支付集成和管理后台接口。

## 🚀 项目概述

Dujiao-Next 是一个开源的自动发卡/数字商品交易平台后端，支持多种支付渠道、多语言、自动/手动发货、钱包系统、推广返利等功能。适用于数字商品（卡密、充值码、软件授权等）的在线销售场景。

## 📋 技术栈

| 技术 | 说明 |
|------|------|
| **Go** | 主要开发语言 |
| **Gin** | HTTP Web 框架 |
| **GORM** | ORM 数据库操作 |
| **SQLite / PostgreSQL** | 数据库支持 |
| **Redis** | 缓存与消息队列 |
| **Asynq** | 异步任务队列（基于 Redis） |
| **Casbin** | RBAC 权限管理 |
| **JWT** | 用户/管理员认证 |
| **Zap** | 结构化日志 |
| **Viper** | 配置管理 |

## 🏗️ 项目架构

```
dujiao-next/
├── cmd/
│   ├── server/main.go         # API 服务入口
│   └── seed/main.go           # 数据库种子数据工具
├── internal/
│   ├── app/                   # 应用启动与生命周期管理
│   ├── authz/                 # RBAC 授权（Casbin）
│   ├── cache/                 # Redis 缓存层
│   ├── config/                # 配置文件加载与解析
│   ├── constants/             # 全局常量定义
│   ├── http/                  # HTTP 处理器（管理端/用户端/公共）
│   ├── i18n/                  # 国际化（zh-CN, zh-TW, en-US）
│   ├── logger/                # 结构化日志（Zap + 日志轮转）
│   ├── models/                # 数据模型（GORM）
│   ├── payment/               # 支付集成（7 种支付方式）
│   ├── provider/              # 依赖注入容器
│   ├── queue/                 # 异步任务定义
│   ├── repository/            # 数据访问层
│   ├── router/                # 路由与中间件
│   ├── service/               # 业务逻辑层
│   └── worker/                # 后台任务处理器
├── config.yml.example         # 配置文件模板
├── Dockerfile                 # Docker 构建文件
└── .goreleaser.yaml           # 多平台发布配置
```

## ✨ 核心功能

### 支付集成
- **Stripe** — 国际信用卡支付（Webhook 签名验证）
- **PayPal** — 国际支付
- **支付宝（Alipay）** — 扫码/网页/WAP 支付（RSA2 签名）
- **微信支付（WeChat Pay）** — Native/JSAPI/H5 支付
- **易支付（Epay）** — 聚合支付（v1 MD5 / v2 RSA）
- **BEpusdt** — USDT/TRC20 加密货币支付
- **TokenPay** — 加密货币支付

### 用户系统
- 邮箱注册/登录（验证码）
- JWT 认证（管理员/用户双 Token 系统）
- Telegram OAuth 登录
- 密码策略（长度、大小写、数字、特殊字符）
- 登录频率限制与账号保护

### 商品管理
- 分类管理
- 商品（支持多 SKU）
- 自动发货（卡密库存）
- 手动发货（人工处理）
- 库存告警

### 订单与钱包
- 完整订单流程（创建 → 支付 → 发货 → 完成）
- 钱包充值与余额支付
- 礼品卡系统
- 购物车

### 管理后台
- RBAC 角色权限管理（Casbin）
- 仪表盘与数据统计
- 系统设置管理
- 通知中心（邮件 + Telegram）
- 推广返利系统
- 优惠券系统

### 其他特性
- 多语言支持（简体中文、繁体中文、英文）
- 结构化日志与日志轮转
- CORS 跨域配置
- 请求 ID 追踪
- 优雅关闭（SIGINT/SIGTERM）

## 🚀 快速开始

### 前置条件

- Go 1.25+
- Redis（用于缓存和队列，可选但推荐）

### 安装与运行

```bash
# 克隆项目
git clone https://github.com/dujiao-next/dujiao-next.git
cd dujiao-next

# 复制配置文件
cp config.yml.example config.yml

# 编辑配置（至少修改 JWT secret）
vim config.yml

# 下载依赖
go mod tidy

# 运行服务
go run cmd/server/main.go
```

### 启动模式

```bash
# 同时启动 API 和 Worker（默认）
go run cmd/server/main.go -mode all

# 仅启动 API 服务
go run cmd/server/main.go -mode api

# 仅启动 Worker 后台任务
go run cmd/server/main.go -mode worker
```

### 种子数据（开发/演示用）

```bash
go run cmd/seed/main.go
```

### Docker 部署（快速启动）

```bash
# 构建镜像
docker build -t dujiao-next .

# 运行容器
docker run -d \
  -p 8080:8080 \
  -v ./config.yml:/app/config.yml:ro \
  -v ./db:/app/db \
  -v ./uploads:/app/uploads \
  -v ./logs:/app/logs \
  dujiao-next
```

### Docker Compose 部署（推荐）

项目提供了符合 CIS Docker Benchmark 和 PCI-DSS 安全基准的 `docker-compose.yml`：

```bash
# 1. 准备配置文件
cp config.yml.example config.yml
vim config.yml  # 修改 JWT secret、数据库等配置

# 2. 生成 Redis 密码并写入 .env（需与 config.yml 中 redis.password 一致）
echo "DJ_REDIS_PASSWORD=$(openssl rand -base64 32)" >> .env

# 3. 启动服务
docker compose up -d

# 4. 查看日志（首次启动时管理员凭据会显示在这里）
docker compose logs -f dujiao-api

# 5. 停止服务
docker compose down
```

## ⚙️ 配置说明

配置文件为 `config.yml`，参考 `config.yml.example`。支持环境变量覆盖（将配置路径中的 `.` 替换为 `_` 并大写，例如 `SERVER_PORT=8080`）。

### 关键配置项

| 配置项 | 说明 | 注意事项 |
|--------|------|----------|
| `server.mode` | 运行模式（debug/release） | 生产环境请设为 `release` |
| `jwt.secret` | JWT 签名密钥 | **生产环境必须 ≥ 32 字符随机字符串** |
| `user_jwt.secret` | 用户 JWT 签名密钥 | **生产环境必须 ≥ 32 字符随机字符串** |
| `database.driver` | 数据库类型（sqlite/postgres） | SQLite 适合开发，PostgreSQL 推荐生产 |
| `redis.enabled` | 是否启用 Redis | 生产环境推荐启用 |

### 默认管理员

首次启动时，如果数据库中没有管理员账号，系统会自动创建一个：

**方式 1：通过环境变量指定凭据（推荐用于生产部署）**
```bash
export DJ_DEFAULT_ADMIN_USERNAME=admin
export DJ_DEFAULT_ADMIN_PASSWORD=your-strong-password-here
```

**方式 2：通过配置文件指定凭据**
```yaml
bootstrap:
  default_admin_username: "admin"
  default_admin_password: "your-strong-password-here"
```

**方式 3：自动生成随机密码（推荐首次体验）**

如果不指定密码，系统会自动生成 24 位随机密码，并以醒目的格式打印到终端（stderr）：

```
╔══════════════════════════════════════════════════════════════╗
║           ⚠️  默认管理员账号已自动创建                        ║
║           ⚠️  Default admin account created                  ║
╠══════════════════════════════════════════════════════════════╣
║                                                              ║
║   用户名 / Username : admin                                  ║
║   密  码 / Password : a1b2c3d4e5f6a1b2c3d4e5f6              ║
║                                                              ║
║   ⚠️  请立即登录后台修改此密码！                              ║
║   ⚠️  Please change this password immediately!               ║
║                                                              ║
╚══════════════════════════════════════════════════════════════╝
```

> **Docker 用户**：使用 `docker compose logs dujiao-api` 或 `docker logs <容器名>` 查看此输出。

## 🔒 安全建议

1. **生产环境必须更改 JWT secret**，长度 ≥ 32 字符
2. **配置文件中的敏感信息**（密码、Token）建议使用环境变量
3. **CORS 配置** 请指定具体域名，避免使用 `*`
4. **数据库** 生产环境推荐使用 PostgreSQL
5. **HTTPS** 请在反向代理（Nginx）层配置 TLS
6. **首次登录后立即修改管理员密码**

> 📄 详细安全审查报告请参见 [SECURITY.md](./SECURITY.md)

## 🐳 Docker 安全部署指南（CIS / PCI-DSS）

本项目的 Dockerfile 和 docker-compose.yml 已按照以下安全基准进行加固：

### CIS Docker Benchmark 合规项

| CIS 编号 | 要求 | 实现状态 |
|-----------|------|----------|
| 4.1 | 使用受信任的最小化基础镜像 | ✅ Alpine 3.21 + 仅安装 ca-certificates, tzdata |
| 4.2 | 容器不以 root 用户运行 | ✅ 使用 `appuser:appgroup` 非特权用户 |
| 4.6 | 添加 HEALTHCHECK 指令 | ✅ 30s 间隔检查 `/health` |
| 5.2 | 限制容器资源（CPU/内存） | ✅ docker-compose 中设置 limits |
| 5.3 | 丢弃不需要的 Linux 能力 | ✅ `cap_drop: ALL` |
| 5.12 | 禁止提权 | ✅ `no-new-privileges:true` |
| 5.25 | 使用只读根文件系统（Redis） | ✅ Redis 容器启用 `read_only: true` |

### PCI-DSS 合规项

| PCI-DSS 编号 | 要求 | 实现状态 |
|---------------|------|----------|
| 2.2.1 | 仅启用必要服务和功能 | ✅ Redis 禁用 FLUSHALL/DEBUG 命令 |
| 6.4.2 | 开发/测试与生产环境分离 | ✅ `server.mode: release` 强制要求强 JWT secret |
| 8.6 | 应用和系统使用唯一身份 | ✅ 非 root 用户运行 + 随机密码生成 |

### 生产部署清单

```bash
# 1. 生成强随机 JWT Secret
openssl rand -base64 48
# 输出示例: xK9mL2pQ...（复制到 config.yml 的 jwt.secret 和 user_jwt.secret）

# 2. 准备配置文件
cp config.yml.example config.yml

# 3. 编辑生产配置（必须修改的项）
#    - server.mode: release
#    - jwt.secret: <上面生成的强密钥>
#    - user_jwt.secret: <另一个强密钥>
#    - redis.password: <生成的 Redis 密码>
#    - database.driver: postgres（推荐）
#    - cors.allowed_origins: ["https://your-domain.com"]
vim config.yml

# 4. 创建 .env 文件（Docker Compose 使用）
cat > .env <<EOF
DJ_REDIS_PASSWORD=$(openssl rand -base64 32)
DJ_DEFAULT_ADMIN_PASSWORD=$(openssl rand -base64 24)
EOF
cat .env  # 请妥善保存此文件中的密码

# 5. 启动服务
docker compose up -d

# 6. 验证健康检查
curl http://localhost:8080/health

# 7. 查看管理员凭据（如使用随机密码，凭据在日志中）
docker compose logs dujiao-api | head -30

# 8. 配置反向代理 + TLS（Nginx 示例）
#    server {
#        listen 443 ssl;
#        server_name api.your-domain.com;
#        ssl_certificate /path/to/cert.pem;
#        ssl_certificate_key /path/to/key.pem;
#        location / {
#            proxy_pass http://127.0.0.1:8080;
#            proxy_set_header Host $host;
#            proxy_set_header X-Real-IP $remote_addr;
#            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
#            proxy_set_header X-Forwarded-Proto $scheme;
#        }
#    }
```

## 📚 API 概览

| 端点 | 说明 |
|------|------|
| `GET /health` | 健康检查 |
| `/api/v1/public/*` | 公共接口（商品、支付、订单查询） |
| `/api/v1/user/*` | 用户接口（需 JWT 认证） |
| `/api/v1/admin/*` | 管理端接口（需管理员 JWT + RBAC） |

## 🧪 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包测试
go test ./internal/service/...
go test ./internal/payment/...
```

## 📦 发布

项目使用 [GoReleaser](https://github.com/goreleaser/goreleaser) 进行多平台构建和发布：

```bash
goreleaser release --clean
```

支持的平台：Linux / macOS / Windows（amd64 / arm64）

## 🔗 在线文档

- https://dujiao-next.com

## 📄 License

参见 [LICENSE](./LICENSE) 文件。
