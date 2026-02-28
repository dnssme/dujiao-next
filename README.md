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

### Docker 部署

```bash
# 构建镜像
docker build -t dujiao-next .

# 运行容器
docker run -d \
  -p 8080:8080 \
  -v ./config.yml:/app/config.yml \
  -v ./db:/app/db \
  -v ./uploads:/app/uploads \
  -v ./logs:/app/logs \
  dujiao-next
```

## ⚙️ 配置说明

配置文件为 `config.yml`，参考 `config.yml.example`。支持环境变量覆盖。

### 关键配置项

| 配置项 | 说明 | 注意事项 |
|--------|------|----------|
| `server.mode` | 运行模式（debug/release） | 生产环境请设为 `release` |
| `jwt.secret` | JWT 签名密钥 | **生产环境必须 ≥ 32 字符随机字符串** |
| `user_jwt.secret` | 用户 JWT 签名密钥 | **生产环境必须 ≥ 32 字符随机字符串** |
| `database.driver` | 数据库类型（sqlite/postgres） | SQLite 适合开发，PostgreSQL 推荐生产 |
| `redis.enabled` | 是否启用 Redis | 生产环境推荐启用 |

### 默认管理员

首次启动时，系统会自动创建管理员账号：

- 通过环境变量指定：`DJ_DEFAULT_ADMIN_USERNAME` / `DJ_DEFAULT_ADMIN_PASSWORD`
- 通过配置文件指定：`bootstrap.default_admin_username` / `bootstrap.default_admin_password`
- 如未指定密码，系统会**自动生成随机密码**并输出到日志

## 🔒 安全建议

1. **生产环境必须更改 JWT secret**，长度 ≥ 32 字符
2. **配置文件中的敏感信息**（密码、Token）建议使用环境变量
3. **CORS 配置** 请指定具体域名，避免使用 `*`
4. **数据库** 生产环境推荐使用 PostgreSQL
5. **HTTPS** 请在反向代理（Nginx）层配置 TLS

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
