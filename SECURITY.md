# 安全审查报告 / Security Audit Report

> 本文件记录了对 Dujiao-Next API 后端的安全审查结果，以及 CIS Docker Benchmark 和 PCI-DSS 合规状态。

## 📋 审查范围

| 审查项 | 文件/组件 |
|--------|-----------|
| 认证与授权 | `internal/router/middleware.go`, `internal/service/auth_service.go`, `internal/authz/` |
| 密码处理 | `internal/models/init.go`, `internal/service/auth_service.go` |
| 数据库安全 | `internal/models/db.go`, `internal/repository/` |
| 支付安全 | `internal/payment/` (Alipay, Stripe, WeChat, PayPal, Epay, TokenPay, BEpusdt) |
| 文件上传 | `internal/service/upload_service.go` |
| 输入验证 | `internal/http/handlers/` |
| 限流防护 | `internal/router/rate_limit.go` |
| Docker 部署 | `Dockerfile`, `docker-compose.yml` |
| 配置安全 | `internal/config/config.go`, `config.yml.example` |

## ✅ 审查结果摘要

**整体风险等级：低（LOW）**

未发现关键安全漏洞。代码采用了业界最佳安全实践。

---

## 🔐 认证与授权

### JWT 认证 ✅ 安全

- **管理员 JWT**：使用 HS256 签名算法，支持 Token 版本控制和失效时间校验
- **用户 JWT**：独立密钥体系，支持「记住我」功能，包含用户状态校验
- **Token 撤销**：通过 `TokenVersion` 和 `TokenInvalidBefore` 实现精确撤销
- **Redis 缓存**：认证状态缓存加速，降低数据库压力

### RBAC 权限管理 ✅ 安全

- 使用 Casbin 实现角色基础访问控制（RBAC）
- 超级管理员（`is_super`）旁路独立于 Casbin 策略
- 权限变更有审计日志记录

---

## 🔑 密码处理

### 密码存储 ✅ 安全

- 使用 `bcrypt`（`golang.org/x/crypto/bcrypt`）进行密码哈希
- 采用 `bcrypt.DefaultCost`（当前为 10）
- 密码从不以明文形式存储或记录在日志中

### 密码策略 ✅ 安全

- 可配置的最小长度（默认 8 位）
- 支持要求大写、小写、数字、特殊字符
- 通过 `security.password_policy` 配置项控制

### 随机密码生成 ✅ 安全

- 使用 `crypto/rand` 生成密码（非 `math/rand`）
- 默认生成 24 位十六进制字符密码
- 生成的密码通过 stderr 输出到终端，不写入结构化日志文件

---

## 🗃️ 数据库安全

### SQL 注入防护 ✅ 安全

- 全部使用 GORM ORM 框架的参数化查询
- 未发现任何原始 SQL 字符串拼接
- 所有 `Where()` 调用均使用占位符 `?`

### 数据库连接 ✅ 安全

- 支持 SQLite（开发）和 PostgreSQL（生产）
- 连接池配置可调（最大连接数、空闲连接数、生命周期）

---

## 💳 支付安全

### 签名验证 ✅ 安全

| 支付方式 | 签名算法 | 验证方式 |
|----------|----------|----------|
| Alipay（支付宝） | RSA2 (SHA256WithRSA) | 公钥验签 + 商户归属校验 |
| WeChat Pay（微信支付） | 官方 SDK | 内置签名验证 |
| Stripe | HMAC-SHA256 | Webhook 签名校验 |
| PayPal | 官方 API | Webhook 事件验证 |
| Epay（易支付） | MD5 / RSA | 签名验证 + 常量时间比较 |
| BEpusdt | HMAC | 签名校验 |
| TokenPay | MD5 HMAC | 签名校验 |

### 回调安全 ✅ 安全

- 所有支付回调均验证签名后才处理
- 订单状态使用幂等校验，防止重复处理
- 回调请求体在日志中做截断处理，避免泄露敏感信息

---

## 📁 文件上传安全

### 上传防护 ✅ 安全

- 文件大小限制（默认 10MB）
- 文件扩展名白名单（`.jpg`, `.jpeg`, `.png`, `.gif`, `.webp`）
- MIME 类型检测和校验
- 图片尺寸限制（最大 4096x4096）
- 使用 UUID 生成安全文件名，防止路径遍历

---

## 🛡️ 请求限流

### 登录限流 ✅ 安全

- 基于 Redis 的滑动窗口限流
- 使用 Lua 脚本实现原子操作
- 支持 IP + 特定字段联合限流
- 默认：5 分钟内最多 5 次，超限后封禁 15 分钟

---

## ⚙️ 配置安全

### JWT Secret 校验 ✅ 安全

- `release` 模式下强制要求 ≥32 字符的 JWT Secret
- 自动检测常见弱密钥（`change-me`、`your-secret-key` 等）
- 开发模式下仅警告，生产模式下直接拒绝启动

### 敏感信息处理 ✅ 安全

- 支持环境变量覆盖配置文件中的敏感值
- 管理员密码仅在首次创建时通过 stderr 输出

---

## 🐳 Docker 安全（CIS Docker Benchmark）

### Dockerfile 合规状态

| CIS 编号 | 要求 | 状态 | 实现说明 |
|-----------|------|------|----------|
| 4.1 | 使用受信任的最小化基础镜像 | ✅ | Alpine 3.21 + 仅安装 ca-certificates, tzdata, wget |
| 4.2 | 不以 root 用户运行容器 | ✅ | 使用 `appuser:appgroup` 非特权用户 |
| 4.3 | 不安装不必要的软件包 | ✅ | 最小化安装，构建阶段独立 |
| 4.6 | 文件系统权限最小化 | ✅ | 目录 750，二进制文件 550 |
| 4.7 | 不在 Dockerfile 中存储密钥 | ✅ | 密钥通过环境变量/配置文件注入 |
| 4.9 | 使用 COPY 而非 ADD | ✅ | 仅使用 COPY 指令 |
| 4.10 | 多阶段构建 | ✅ | builder → runtime 两阶段 |

### docker-compose.yml 合规状态

| CIS 编号 | 要求 | 状态 | 实现说明 |
|-----------|------|------|----------|
| 5.2 | 限制容器资源 | ✅ | CPU 和内存限制 |
| 5.3 | 丢弃不需要的 Linux 能力 | ✅ | `cap_drop: ALL` |
| 5.4 | 不使用特权容器 | ✅ | `privileged: false` |
| 5.12 | 禁止容器获取新权限 | ✅ | `no-new-privileges:true` |
| 5.14 | 限制容器内网络 | ✅ | 使用独立 `internal` bridge 网络 |
| 5.25 | 只读根文件系统 | ✅ | Redis 容器启用 `read_only: true` |
| 5.26 | 容器健康检查 | ✅ | Redis 和 API 均配置 HEALTHCHECK |

---

## 📜 PCI-DSS 合规状态

| PCI-DSS 编号 | 要求 | 状态 | 实现说明 |
|---------------|------|------|----------|
| 2.2.1 | 仅启用必要服务 | ✅ | Redis 禁用 FLUSHALL/DEBUG；最小化 Alpine 镜像 |
| 2.2.2 | 不使用默认密码 | ✅ | JWT secret 强制检查；管理员密码随机生成 |
| 6.3.1 | 代码审查 | ✅ | 本审查报告 |
| 6.5.1 | 防止注入漏洞 | ✅ | GORM 参数化查询，无 SQL 拼接 |
| 6.5.3 | 安全的加密存储 | ✅ | bcrypt 密码哈希 |
| 6.5.7 | XSS 防护 | ✅ | 纯 JSON API，不输出 HTML |
| 6.5.10 | 防止缓冲区溢出/DoS | ✅ | HTTP 超时设置（Read/Write/Header/Idle） |
| 8.2.1 | 密码认证 | ✅ | Redis 启用密码认证 |
| 8.6 | 唯一标识 | ✅ | 非 root 用户运行 + UUID 文件名 |
| 10.1 | 审计日志 | ✅ | 结构化日志 + 请求 ID 追踪 |
| 10.3 | 日志内容 | ✅ | 包含用户 ID、IP、时间戳、操作路径 |

---

## 🔧 改进建议

以下为非必须但可增强安全性的建议：

1. **生产环境建议**
   - 使用 PostgreSQL 替代 SQLite
   - 在 Nginx 反向代理层配置 TLS（HTTPS）
   - 配置具体的 CORS 允许域名，不使用 `*`

2. **可选加固**
   - 定期轮换 JWT Secret
   - 添加 IP 白名单限制管理接口访问

---

## 🏗️ 部署安全清单

### 首次部署必做事项

```bash
# 1. 生成强随机 JWT Secret（≥32 字符）
openssl rand -base64 48

# 2. 生成 Redis 密码
openssl rand -base64 32

# 3. 准备配置文件
cp config.yml.example config.yml

# 4. 编辑配置（必须修改项标记 ⚠️）
#    ⚠️ server.mode: release
#    ⚠️ jwt.secret: <生成的随机密钥>
#    ⚠️ user_jwt.secret: <另一个随机密钥>
#    ⚠️ redis.password: <生成的 Redis 密码>
#    ⚠️ cors.allowed_origins: ["https://your-domain.com"]
#    推荐 database.driver: postgres

# 5. 创建 .env 文件（docker compose 使用）
cat > .env <<EOF
DJ_REDIS_PASSWORD=<生成的 Redis 密码>
DJ_DEFAULT_ADMIN_PASSWORD=<你的管理员密码>
EOF

# 6. 启动服务
docker compose up -d

# 7. 查看管理员凭据（如未指定密码，随机密码在此显示）
docker compose logs dujiao-api

# 8. 配置 TLS 反向代理（Nginx）
```

---

## 📝 审查声明

- 审查日期：2026-02-28
- 审查范围：全部 Go 源代码、Dockerfile、docker-compose.yml、配置文件
- 审查方法：人工代码审查 + 自动化静态分析（CodeQL）
- 审查结论：未发现关键安全漏洞，代码质量良好

---

## ⚠️ 前端兼容性说明

本仓库（`dujiao-next/dujiao-next`）是**纯 JSON API 后端**。前端页面和管理页面位于独立仓库：

- 用户前端：[dujiao-next/user](https://github.com/dujiao-next/user)
- 管理后台：[dujiao-next/admin](https://github.com/dujiao-next/admin)

本次安全加固的所有修改**不会影响前端页面功能**，因为：

1. **未修改任何 API 路由**（`/api/v1/*` 所有端点不变）
2. **未修改任何请求/响应格式**（JSON 结构不变）
3. **未修改任何认证流程**（JWT 认证逻辑不变）
4. **仅修改了部署层面配置**（Dockerfile、docker-compose.yml）
5. **仅改进了管理员密码的显示方式**（从日志改为 stderr 输出）
