# Dujiao-Next 部署指南 / Deployment Guide

> 本指南仅涵盖 Dujiao-Next 程序本身的部署。
> NGINX、Redis 等外部依赖请自行部署，本文档提供配置要求供参考。

---

## 目录

- [架构概览](#架构概览)
- [外部依赖要求](#外部依赖要求)
- [API 后端部署（Docker）](#api-后端部署docker)
- [User 前端部署](#user-前端部署)
- [Admin 前端部署](#admin-前端部署)
- [NGINX 参考配置](#nginx-参考配置)
- [OWASP CRS 防火墙规则](#owasp-crs-防火墙规则)
- [安全加固清单](#安全加固清单)
- [运维操作](#运维操作)
- [故障排查](#故障排查)

---

## 架构概览

```
                 ┌────────────────────────────────────┐
                 │     你自己部署的 NGINX              │
                 │  • TLS 终止                         │
                 │  • OWASP CRS WAF 防火墙             │
                 │  • Rate Limiting 限流               │
                 │  • Security Headers 安全头           │
                 └──┬──────────────┬─────────────┬───┘
                    │              │             │
           ┌────────▼──┐   ┌──────▼─────┐  ┌───▼───────┐
           │  User 前端 │   │ Admin 前端  │  │ API 后端   │
           │ 静态文件    │   │ 静态文件    │  │ Docker     │
           │ /var/www/  │   │ /var/www/  │  │ :8080      │
           └────────────┘   └────────────┘  └──────┬────┘
                                                    │
                                              ┌─────▼─────┐
                                              │   Redis    │
                                              │ 你自己部署  │
                                              └───────────┘
```

---

## 外部依赖要求

以下组件需要你自行安装和部署，不包含在本程序的 Docker 中：

### Redis

| 项目 | 要求 |
|------|------|
| 版本 | Redis 6.0+ (推荐 7.x) |
| 认证 | **必须设置密码** (`requirepass`) |
| 用途 | 缓存 + 异步任务队列 |
| 安全建议 | 仅监听 127.0.0.1；禁用危险命令（FLUSHALL/DEBUG） |

Redis 安全配置示例：
```conf
# /etc/redis/redis.conf
bind 127.0.0.1
requirepass <你的强密码>
maxmemory 256mb
maxmemory-policy allkeys-lru
rename-command FLUSHALL ""
rename-command FLUSHDB ""
rename-command DEBUG ""
```

### NGINX

| 项目 | 要求 |
|------|------|
| 版本 | NGINX 1.20+ |
| 用途 | 反向代理 + TLS 终止 + WAF |
| 可选 | ModSecurity + OWASP CRS |

参考配置文件已提供：`nginx/nginx.conf`

### 数据库（可选）

| 模式 | 说明 |
|------|------|
| SQLite（默认） | 无需额外安装，适合小规模 |
| PostgreSQL | 生产环境推荐，需自行安装 |

---

## API 后端部署（Docker）

### 步骤 1: 准备配置文件

```bash
# 克隆仓库
git clone https://github.com/dujiao-next/dujiao-next.git
cd dujiao-next

# 创建配置文件
cp config.yml.example config.yml
cp .env.example .env
```

### 步骤 2: 修改 config.yml

```bash
vim config.yml
```

**必须修改的配置项** ⚠️：

```yaml
server:
  mode: release             # ⚠️ 生产环境必须改为 release

jwt:
  secret: "<≥32字符随机字符串>"  # ⚠️ 必须修改！生成方法见下方

user_jwt:
  secret: "<另一个≥32字符随机字符串>"  # ⚠️ 必须修改！

redis:
  host: 127.0.0.1           # ⚠️ 改为你的 Redis 地址
  port: 6379
  password: "<你的Redis密码>"  # ⚠️ 必须填写

queue:
  host: 127.0.0.1           # ⚠️ 与 redis.host 相同
  port: 6379
  password: "<你的Redis密码>"  # ⚠️ 与 redis.password 相同

cors:
  allowed_origins:
    - "https://your-domain.com"        # ⚠️ 改为你的域名
    - "https://admin.your-domain.com"  # ⚠️ 管理后台域名
```

**生成随机密钥：**

```bash
# JWT Secret（≥32 字符）
openssl rand -base64 48

# Redis 密码
openssl rand -base64 32
```

### 步骤 3: 构建并启动 API

```bash
# 构建 Docker 镜像
docker compose build

# 启动服务（后台运行）
docker compose up -d

# 查看日志（首次启动时管理员凭据会在这里显示）
docker compose logs -f dujiao-api
```

### 步骤 4: 验证 API 运行

```bash
# 健康检查
curl http://127.0.0.1:8080/health

# 获取公共配置
curl http://127.0.0.1:8080/api/v1/public/config
```

### Docker 合规说明

API 容器遵循 CIS Docker Benchmark 和 PCI-DSS 安全标准：

| CIS 编号 | 要求 | 实现方式 |
|-----------|------|----------|
| 4.1 | 受信任最小化基础镜像 | Alpine 3.21 |
| 4.2 | 非 root 用户 | `appuser:appgroup` |
| 4.6 | 最小文件权限 | 目录 750，二进制 550 |
| 4.10 | 多阶段构建 | builder → runtime |
| 5.2 | 资源限制 | CPU 2.0 / 内存 512M |
| 5.3 | 最小 Linux 能力 | `cap_drop: ALL` |
| 5.12 | 禁止新权限 | `no-new-privileges` |
| 5.26 | 健康检查 | `/health` 端点 |
| 6.1 | HEALTHCHECK 指令 | Dockerfile 内置 |

---

## User 前端部署

User 前端是 Vue 3 SPA 应用，构建后为纯静态文件。

### 方法 A: 直接构建部署（推荐）

```bash
# 进入前端目录
cd user

# 安装依赖
npm ci

# 配置 API 地址（可选，默认使用同域名）
# export VITE_API_BASE_URL=https://your-domain.com

# 构建
npm run build

# 部署到 NGINX 静态目录
cp -r dist/* /var/www/user/dist/
```

### 方法 B: Docker 构建后提取文件

```bash
# 构建镜像
cd user
docker build -t dujiao-user:latest .

# 提取静态文件
docker create --name tmp-user dujiao-user:latest
docker cp tmp-user:/usr/share/nginx/html /var/www/user/dist
docker rm tmp-user
```

### NGINX 配置

User 前端作为 SPA 应用，NGINX 需要配置回退到 `index.html`：

```nginx
root /var/www/user/dist;
index index.html;

location / {
    try_files $uri $uri/ /index.html;
}

location ^~ /assets/ {
    expires 1y;
    add_header Cache-Control "public, max-age=31536000, immutable" always;
}
```

完整的 NGINX 配置参考 `nginx/nginx.conf`。

---

## Admin 前端部署

Admin 前端同样是 Vue 3 SPA 应用，部署方法与 User 相同。

### 构建

```bash
cd admin

npm ci

# 配置 API 地址（可选）
# export VITE_API_BASE_URL=https://your-domain.com

npm run build

cp -r dist/* /var/www/admin/dist/
```

### 安全建议

管理后台应额外加固：

1. **使用独立端口或子域名**（如 `admin.your-domain.com` 或 `:81`）
2. **IP 白名单** — 在 NGINX 中限制只有管理员 IP 可访问
3. **更严格的 CSP** — 管理后台不需要 Telegram 等外部脚本
4. **X-Frame-Options: DENY** — 防止 Clickjacking

示例 NGINX IP 限制：
```nginx
server {
    # ...管理后台配置...
    allow 你的办公IP/32;
    allow 10.0.0.0/8;    # 内网
    deny all;
}
```

---

## NGINX 参考配置

项目提供了完整的 NGINX 参考配置文件：

```
nginx/nginx.conf                     # NGINX 反向代理配置（含安全头+限流）
nginx/modsecurity/dujiao-crs-rules.conf  # OWASP CRS 自定义规则
```

### 安装步骤

```bash
# 1. 复制配置
cp nginx/nginx.conf /etc/nginx/conf.d/dujiao.conf

# 2. 修改配置中的域名和路径
vim /etc/nginx/conf.d/dujiao.conf
#   - 修改 server_name
#   - 修改 root 路径
#   - 修改 upstream 中的 API 地址
#   - 配置 TLS 证书路径

# 3. 测试并重载
nginx -t && nginx -s reload
```

### NGINX 安全特性

配置文件包含以下安全措施：

| 特性 | 说明 |
|------|------|
| Rate Limiting | API 20r/s, 登录 5r/m, 静态 50r/s |
| Security Headers | X-Content-Type-Options, X-Frame-Options, CSP 等 |
| TLS 配置 | TLSv1.2/1.3, 安全密码套件（需取消注释） |
| 路径屏蔽 | 屏蔽 .git, .env, .sql 等敏感文件 |
| IP 白名单 | 管理后台可配置 IP 限制 |

---

## OWASP CRS 防火墙规则

项目提供了针对 Dujiao-Next 的 OWASP CRS 追加规则文件：
`nginx/modsecurity/dujiao-crs-rules.conf`

### 使用方法

前提：你已经安装并配置好了 ModSecurity + OWASP CRS。

```bash
# 1. 复制规则文件
cp nginx/modsecurity/dujiao-crs-rules.conf /etc/nginx/modsecurity/

# 2. 在 modsecurity.conf 中 Include（在默认 CRS 规则之后）
#    示例加载顺序：
#    Include /etc/nginx/modsecurity/crs/crs-setup.conf
#    Include /etc/nginx/modsecurity/crs/rules/*.conf
#    Include /etc/nginx/modsecurity/dujiao-crs-rules.conf  ← 添加这行
```

### 规则说明

| 规则 ID | 说明 | 作用 |
|---------|------|------|
| 10001-10002 | 支付回调豁免 | 避免支付宝/微信等回调被误拦 |
| 10003 | 文件上传豁免 | 允许管理后台图片上传 |
| 10004 | 富文本内容豁免 | 允许管理后台保存 HTML 内容 |
| 10005 | JSON 请求豁免 | 允许 API 的 JSON 请求体 |
| 10006 | CSV 导入豁免 | 允许卡密批量导入 |
| 10007 | JWT Token 豁免 | 避免 Bearer Token 触发 SQLi |
| 10008 | 验证码接口豁免 | 验证码特殊字符不触发规则 |
| 10009 | 优惠券规则豁免 | 优惠券 JSON 配置不触发规则 |
| 10010 | Telegram OAuth 豁免 | Telegram 登录签名数据 |
| 10011 | 推广链接豁免 | 推广 referrer 数据 |

### 推荐的 CRS 全局配置

在你的 `crs-setup.conf` 中设置：

```
Paranoia Level: 2（推荐）
异常评分阈值: 入站 10, 出站 5
```

---

## 安全加固清单

### 部署后必做 ✅

- [ ] `config.yml` 中 `server.mode` 设为 `release`
- [ ] 设置强 JWT Secret（两个，各 ≥32 字符随机字符串）
- [ ] 设置 Redis 密码
- [ ] 修改 CORS `allowed_origins` 为你的具体域名（不要用 `*`）
- [ ] NGINX 配置 TLS/HTTPS（Let's Encrypt 或购买证书）
- [ ] NGINX 配置 `server_name` 为你的域名
- [ ] 查看并保存首次启动时日志中的管理员密码

### 建议加固 🔒

- [ ] 管理后台限制 IP 白名单
- [ ] 启用 OWASP ModSecurity CRS WAF
- [ ] 生产环境使用 PostgreSQL 替代 SQLite
- [ ] 定期备份数据库和上传文件
- [ ] 定期更新 Docker 镜像
- [ ] 定期轮换 JWT Secret 和 Redis 密码
- [ ] 配置日志轮转和监控告警

---

## 运维操作

### 日常操作

```bash
# 查看 API 容器状态
docker compose ps

# 查看 API 日志
docker compose logs -f dujiao-api

# 重启 API
docker compose restart dujiao-api

# 更新代码并重新构建
git pull
docker compose build
docker compose up -d

# 查看资源使用
docker stats dujiao-api --no-stream
```

### 备份数据

```bash
# 备份数据库（SQLite）
docker cp dujiao-api:/app/db/dujiao.db ./backup/dujiao.db

# 备份上传文件
docker cp dujiao-api:/app/uploads ./backup/uploads/

# 备份配置
cp config.yml ./backup/
```

### 恢复数据

```bash
# 停止服务
docker compose down

# 恢复数据库
docker cp ./backup/dujiao.db dujiao-api:/app/db/dujiao.db

# 重新启动
docker compose up -d
```

---

## 故障排查

### API 容器启动失败

```bash
# 查看详细日志
docker compose logs dujiao-api

# 常见原因：
# - config.yml 语法错误
# - Redis 连接失败（检查地址和密码）
# - 端口被占用
```

### API 无法连接 Redis

```bash
# 检查 Redis 是否运行
redis-cli -h 127.0.0.1 -a <密码> ping

# 检查 config.yml 中的 Redis 配置
# 注意：如果 API 运行在 Docker 中，Redis 地址不能用 127.0.0.1
# 应该使用宿主机 IP 或 host.docker.internal
```

**Docker 中的 API 连接宿主机 Redis：**

```yaml
# docker-compose.yml 中添加
extra_hosts:
  - "host.docker.internal:host-gateway"
```

```yaml
# config.yml 中使用
redis:
  host: host.docker.internal
```

### 前端显示 502 或 CORS 错误

1. 确认 API 健康：`curl http://127.0.0.1:8080/health`
2. 确认 NGINX 反向代理指向正确地址
3. 检查 `config.yml` 中 CORS 配置包含前端域名

### 管理员忘记密码

```bash
# 方法 1: 查看首次启动日志
docker compose logs dujiao-api | grep -i password

# 方法 2: 删除数据库重新初始化（⚠️ 丢失所有数据）
docker compose down
docker volume rm dujiao_db
docker compose up -d
# 查看新密码
docker compose logs dujiao-api
```

---

## 目录结构

```
dujiao-next/
├── Dockerfile              # API 后端 Docker 构建文件（CIS 加固）
├── docker-compose.yml      # API Docker Compose 编排（仅 API）
├── config.yml.example      # API 配置模板
├── .env.example            # 环境变量模板
├── nginx/
│   ├── nginx.conf          # NGINX 参考配置（你自己部署 NGINX 时使用）
│   └── modsecurity/
│       └── dujiao-crs-rules.conf  # OWASP CRS 自定义规则
├── user/                   # 用户前台（Vue 3）
│   ├── Dockerfile          # 可选：用 Docker 构建前端
│   └── src/                # 前端源码
├── admin/                  # 管理后台（Vue 3）
│   ├── Dockerfile          # 可选：用 Docker 构建前端
│   └── src/                # 后台源码
├── cmd/                    # Go API 入口
├── internal/               # Go API 业务逻辑
├── SECURITY.md             # 安全审查报告
└── DEPLOYMENT.md           # 本文件
```
