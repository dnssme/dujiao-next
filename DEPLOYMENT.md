# 🚀 Dujiao-Next 全栈部署指南 / Full-Stack Deployment Guide

> CIS Docker Benchmark + PCI-DSS + OWASP CRS 合规部署

## 📋 目录 / Table of Contents

- [架构概览](#架构概览)
- [前置条件](#前置条件)
- [快速开始（5 分钟）](#快速开始)
- [生产部署（详细步骤）](#生产部署)
- [TLS/HTTPS 配置](#tlshttps-配置)
- [OWASP WAF 配置](#owasp-waf-配置)
- [安全加固清单](#安全加固清单)
- [运维操作](#运维操作)
- [故障排查](#故障排查)

---

## 架构概览

```
                    ┌─────────────────────────────────────────────┐
                    │              Internet / CDN                  │
                    └──────────────────┬──────────────────────────┘
                                       │
                    ┌──────────────────▼──────────────────────────┐
                    │          NGINX (端口 80/443/81)              │
                    │   • TLS 终止                                 │
                    │   • OWASP CRS WAF 防火墙                     │
                    │   • Rate Limiting 限流                       │
                    │   • Security Headers 安全头                  │
                    └──┬────────────┬────────────┬───────────────┘
                       │            │            │
              ┌────────▼──┐  ┌─────▼────┐  ┌───▼────────┐
              │ User (80)  │  │ Admin(81)│  │ API (8080) │
              │ Vue3 SPA   │  │ Vue3 SPA │  │ Go + Gin   │
              │ 前台页面    │  │ 管理后台  │  │ JSON API   │
              └────────────┘  └──────────┘  └──────┬─────┘
                                                    │
                                              ┌─────▼─────┐
                                              │   Redis    │
                                              │ 缓存+队列  │
                                              └───────────┘
```

**网络隔离：**
- `frontend` 网络: NGINX ↔ User Frontend / Admin Frontend
- `internal` 网络: NGINX ↔ API ↔ Redis
- 前端容器和 Redis **不直接暴露端口**

---

## 前置条件

| 组件 | 最低版本 | 说明 |
|------|----------|------|
| Docker | 24.0+ | `docker --version` |
| Docker Compose | v2.20+ | `docker compose version` |
| 磁盘空间 | 2GB+ | 构建镜像需要空间 |
| 内存 | 2GB+ | 运行所有容器 |

---

## 快速开始

> ⚡ 适合开发测试环境，5 分钟启动全栈

```bash
# 1. 克隆仓库
git clone https://github.com/dujiao-next/dujiao-next.git
cd dujiao-next

# 2. 准备配置文件
cp config.yml.example config.yml
cp .env.example .env

# 3. 生成安全密钥
# 修改 .env 中的 DJ_REDIS_PASSWORD
sed -i "s/CHANGE_ME_GENERATE_RANDOM_PASSWORD/$(openssl rand -base64 32)/" .env

# 4. 修改 config.yml 中的 redis 密码（需与 .env 一致）
# 同时修改 jwt.secret 和 user_jwt.secret 为强随机值

# 5. 启动全栈
docker compose up -d

# 6. 查看管理员凭据
docker compose logs dujiao-api 2>&1 | head -50

# 7. 访问
# 用户前台: http://localhost
# 管理后台: http://localhost:81
# API 健康检查: http://localhost/api/v1/public/config
```

---

## 生产部署

### 步骤 1: 生成所有密钥

```bash
# JWT Secret（≥32 字符）
echo "JWT_SECRET: $(openssl rand -base64 48)"
echo "USER_JWT_SECRET: $(openssl rand -base64 48)"
echo "REDIS_PASSWORD: $(openssl rand -base64 32)"
echo "ADMIN_PASSWORD: $(openssl rand -base64 24)"
```

### 步骤 2: 配置 .env 文件

```bash
cp .env.example .env
vim .env
```

```env
# 使用上面生成的 Redis 密码
DJ_REDIS_PASSWORD=<你的Redis密码>

# 管理员凭据
DJ_DEFAULT_ADMIN_USERNAME=admin
DJ_DEFAULT_ADMIN_PASSWORD=<你的管理员密码>

# 端口（如果有冲突可修改）
DJ_HTTP_PORT=80
DJ_HTTPS_PORT=443
DJ_ADMIN_PORT=81
```

### 步骤 3: 配置 config.yml

```bash
cp config.yml.example config.yml
vim config.yml
```

**必须修改的配置项** ⚠️：

```yaml
server:
  mode: release  # ⚠️ 生产环境必须改为 release

jwt:
  secret: "<生成的JWT密钥，≥32字符>"  # ⚠️ 必须修改

user_jwt:
  secret: "<另一个JWT密钥，≥32字符>"  # ⚠️ 必须修改

redis:
  host: redis       # ⚠️ Docker Compose 内部使用 redis 作为主机名
  port: 6379
  password: "<与.env中的DJ_REDIS_PASSWORD相同>"  # ⚠️ 必须与.env一致

queue:
  host: redis       # ⚠️ 与 redis.host 相同
  port: 6379
  password: "<与.env中的DJ_REDIS_PASSWORD相同>"  # ⚠️ 必须与.env一致

cors:
  allowed_origins:
    - "https://your-domain.com"      # ⚠️ 改为你的域名
    - "https://admin.your-domain.com" # ⚠️ 管理后台域名
```

### 步骤 4: 配置 NGINX

编辑 `nginx/nginx.conf`：

```bash
vim nginx/nginx.conf
```

修改 `server_name`：
```nginx
# User Frontend server block
server_name your-domain.com;

# Admin Frontend server block  
server_name admin.your-domain.com;
```

### 步骤 5: 构建并启动

```bash
# 构建所有镜像
docker compose build

# 启动服务（后台运行）
docker compose up -d

# 查看所有容器状态
docker compose ps

# 查看 API 日志（获取管理员密码）
docker compose logs -f dujiao-api
```

### 步骤 6: 验证部署

```bash
# 健康检查
curl http://localhost/nginx-health
curl http://localhost/api/v1/public/config

# 查看容器资源使用
docker compose stats --no-stream
```

---

## TLS/HTTPS 配置

### 方法 A: Let's Encrypt（推荐）

```bash
# 1. 安装 certbot
apt install certbot

# 2. 获取证书（先停止 nginx 容器或使用 webroot 模式）
docker compose stop nginx
certbot certonly --standalone -d your-domain.com -d admin.your-domain.com
docker compose start nginx

# 3. 复制证书到 nginx/ssl/
mkdir -p nginx/ssl
cp /etc/letsencrypt/live/your-domain.com/fullchain.pem nginx/ssl/
cp /etc/letsencrypt/live/your-domain.com/privkey.pem nginx/ssl/
chmod 600 nginx/ssl/*.pem

# 4. 编辑 nginx/nginx.conf，取消 TLS 相关注释
# 5. 编辑 docker-compose.yml，取消 ssl volume 挂载注释
# 6. 重启 nginx
docker compose restart nginx
```

### 方法 B: 自签名证书（测试用）

```bash
mkdir -p nginx/ssl
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout nginx/ssl/privkey.pem \
  -out nginx/ssl/fullchain.pem \
  -subj "/CN=your-domain.com"
```

---

## OWASP WAF 配置

### 启用 ModSecurity WAF

项目包含了预配置的 OWASP CRS 规则文件 (`nginx/modsecurity/modsecurity.conf`)。

**方法 1: 使用 OWASP 官方 Docker 镜像**

修改 `docker-compose.yml` 中的 nginx 服务：

```yaml
nginx:
  image: owasp/modsecurity-crs:nginx-alpine
  volumes:
    - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
    - ./nginx/modsecurity/modsecurity.conf:/etc/nginx/modsecurity/modsecurity.conf:ro
```

**方法 2: 使用 Cloudflare（替代方案）**

如果使用 Cloudflare 作为 CDN，可以在 Cloudflare Dashboard 中启用 WAF 规则，无需本地 ModSecurity。

### WAF 规则说明

| 规则 ID | 说明 | 作用 |
|---------|------|------|
| 10001-10002 | 支付回调豁免 | 避免支付回调被误拦截 |
| 10003 | 文件上传豁免 | 允许图片上传 |
| 10004 | 富文本内容豁免 | 允许管理后台保存 HTML 内容 |
| 10005 | JSON 请求豁免 | 允许 API JSON 请求体 |
| 10006 | CSV 导入豁免 | 允许卡密批量导入 |
| 10007 | JWT Token 豁免 | 避免 Bearer Token 触发 SQLi 检测 |

Paranoia Level 设置为 **2**（适中），可在 `modsecurity.conf` 中调整：
- Level 1: 最少误报，基础防护
- Level 2: 推荐，平衡安全与兼容
- Level 3: 严格，可能需要额外调整
- Level 4: 最严格，需要大量调整

---

## 安全加固清单

### 部署后必做 ✅

- [ ] 修改 `config.yml` 中 `server.mode` 为 `release`
- [ ] 设置强 JWT Secret（≥32 字符随机字符串）
- [ ] 设置 Redis 密码
- [ ] 设置管理员强密码
- [ ] 配置 CORS 允许的域名（不使用 `*`）
- [ ] 配置 TLS/HTTPS
- [ ] 修改 NGINX 中的 `server_name`

### 建议加固 🔒

- [ ] 限制管理后台 IP 访问（在 `nginx/nginx.conf` 中取消注释 IP 限制）
- [ ] 启用 OWASP ModSecurity WAF
- [ ] 生产环境使用 PostgreSQL 替代 SQLite
- [ ] 配置日志轮转和监控告警
- [ ] 定期更新 Docker 镜像
- [ ] 定期轮换 JWT Secret 和 Redis 密码

### CIS Docker Benchmark 合规状态

| CIS 编号 | 要求 | 状态 |
|-----------|------|------|
| 4.1 | 受信任最小化基础镜像 | ✅ Alpine |
| 4.2 | 非 root 用户 | ✅ appuser |
| 4.6 | 最小文件权限 | ✅ 750/550 |
| 5.2 | 资源限制 | ✅ CPU+内存 |
| 5.3 | 最小 Linux 能力 | ✅ cap_drop ALL |
| 5.12 | 禁止新权限 | ✅ no-new-privileges |
| 5.14 | 网络隔离 | ✅ frontend+internal |
| 5.25 | 只读文件系统 | ✅ Redis+前端 |
| 5.26 | 健康检查 | ✅ 全部容器 |

---

## 运维操作

### 日常操作

```bash
# 查看所有容器状态
docker compose ps

# 查看 API 日志
docker compose logs -f dujiao-api

# 查看 NGINX 日志
docker compose logs -f nginx

# 重启单个服务
docker compose restart dujiao-api

# 更新镜像并重启
docker compose build
docker compose up -d

# 查看资源使用
docker compose stats --no-stream
```

### 备份数据

```bash
# 备份数据库
docker compose exec dujiao-api cp /app/db/dujiao.db /app/db/dujiao.db.bak
docker cp $(docker compose ps -q dujiao-api):/app/db/dujiao.db ./backup/

# 备份上传文件
docker cp $(docker compose ps -q dujiao-api):/app/uploads ./backup/uploads/

# 备份配置
cp config.yml ./backup/
cp .env ./backup/
```

### 更新部署

```bash
# 拉取最新代码
git pull

# 重新构建并更新
docker compose build
docker compose up -d

# 验证
docker compose ps
curl http://localhost/nginx-health
```

---

## 故障排查

### 容器启动失败

```bash
# 查看详细日志
docker compose logs <service-name>

# 检查容器退出原因
docker inspect $(docker compose ps -q <service-name>) | grep -A5 "State"
```

### API 无法连接 Redis

检查 `config.yml` 中 Redis 配置：
```yaml
redis:
  host: redis    # 使用 Docker Compose 服务名
  password: "xxx" # 必须与 .env 中 DJ_REDIS_PASSWORD 一致
```

### 前端显示 502 错误

```bash
# 检查 API 是否健康
docker compose exec nginx wget -qO- http://dujiao-api:8080/health

# 检查 NGINX 日志
docker compose logs nginx | tail -20
```

### 管理后台无法访问

1. 确认管理后台端口：默认 81（`http://server-ip:81`）
2. 确认防火墙允许端口 81
3. 检查 `docker compose ps` 中 dujiao-admin 容器状态

### CORS 错误

修改 `config.yml`：
```yaml
cors:
  allowed_origins:
    - "https://your-domain.com"
    - "https://admin.your-domain.com"
    - "http://localhost"          # 开发环境
    - "http://localhost:81"       # 管理后台开发
```

---

## 📐 目录结构

```
dujiao-next/
├── Dockerfile              # API 后端 Docker 构建文件
├── docker-compose.yml      # 全栈 Docker Compose 编排
├── config.yml.example      # API 配置模板
├── .env.example            # 环境变量模板
├── nginx/
│   ├── nginx.conf          # NGINX 反向代理配置（含安全头+限流）
│   └── modsecurity/
│       └── modsecurity.conf # OWASP CRS WAF 规则配置
├── user/                   # 用户前台（Vue 3）
│   ├── Dockerfile          # 前台 Docker 构建文件（CIS 加固）
│   ├── src/                # 前台源码
│   └── ...
├── admin/                  # 管理后台（Vue 3）
│   ├── Dockerfile          # 后台 Docker 构建文件（CIS 加固）
│   ├── src/                # 后台源码
│   └── ...
├── cmd/                    # Go API 入口
├── internal/               # Go API 业务逻辑
├── SECURITY.md             # 安全审查报告
└── DEPLOYMENT.md           # 本文件
```
