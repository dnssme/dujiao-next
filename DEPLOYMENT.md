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
  host: 127.0.0.1           # ⚠️ 改为你的 Redis 地址（见下方说明）
  port: 6379
  password: "<你的Redis密码>"  # ⚠️ 必须填写

queue:
  host: 127.0.0.1           # ⚠️ 与 redis.host 相同
  port: 6379
  password: "<你的Redis密码>"  # ⚠️ 与 redis.password 相同
```

> **⚠️ Redis 地址注意：** API 运行在 Docker 容器中，`127.0.0.1` 指向容器内部。
> 如果 Redis 运行在宿主机上，请使用 `host.docker.internal`（见[故障排查](#故障排查)章节）。
> 如果 Redis 运行在其他服务器上，填写该服务器的 IP 地址。

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

User 前端是 Vue 3 + Vite + Tailwind SPA 应用，构建后生成纯静态文件（HTML/JS/CSS），
部署到 NGINX 静态目录即可。**不需要 Node.js 运行时。**

### 前置条件

| 组件 | 版本 | 说明 |
|------|------|------|
| Node.js | 20.x+ | 仅构建时需要 |
| npm | 10.x+ | 仅构建时需要 |

### 方法 A: 在服务器上构建部署

```bash
# 进入前端目录
cd user

# 安装依赖（--ignore-scripts 防止安装后脚本执行，安全构建）
npm ci --ignore-scripts

# 配置 API 地址
# 如果前后端在同一域名下（推荐），不需要设置此变量
# 如果 API 在不同域名（如 api.your-domain.com），取消注释：
# export VITE_API_BASE_URL=https://api.your-domain.com

# 构建生产版本
npm run build

# 创建部署目录
sudo mkdir -p /var/www/user/dist

# 部署静态文件
sudo cp -r dist/* /var/www/user/dist/

# 设置安全文件权限
sudo chown -R www-data:www-data /var/www/user/dist
sudo find /var/www/user/dist -type d -exec chmod 755 {} \;
sudo find /var/www/user/dist -type f -exec chmod 644 {} \;
```

### 方法 B: 在本地构建后上传

```bash
# 本地构建
cd user
npm ci --ignore-scripts
npm run build

# 上传到服务器
scp -r dist/* your-server:/var/www/user/dist/

# 在服务器上设置权限
ssh your-server 'sudo chown -R www-data:www-data /var/www/user/dist && sudo find /var/www/user/dist -type d -exec chmod 755 {} \; && sudo find /var/www/user/dist -type f -exec chmod 644 {} \;'
```

### 方法 C: Docker 构建后提取文件

```bash
# 构建镜像
cd user
docker build -t dujiao-user:latest .

# 提取静态文件
docker create --name tmp-user dujiao-user:latest
docker cp tmp-user:/usr/share/nginx/html/. /var/www/user/dist/
docker rm tmp-user

# 设置权限
sudo chown -R www-data:www-data /var/www/user/dist
sudo find /var/www/user/dist -type d -exec chmod 755 {} \;
sudo find /var/www/user/dist -type f -exec chmod 644 {} \;
```

### 构建产物结构

```
/var/www/user/dist/
├── index.html              # SPA 入口
├── assets/
│   ├── index-a1b2c3d4.js   # JS 文件（带 hash，长缓存）
│   ├── index-e5f6g7h8.css  # CSS 文件（带 hash，长缓存）
│   └── ...
└── favicon.ico             # 图标
```

### User 前端安全加固

1. **NGINX 配置 SPA 路由回退**：所有非文件请求回退到 `index.html`
2. **静态资源长缓存**：`/assets/` 目录设置 1 年缓存（文件名含 hash）
3. **安全头**：CSP、X-Frame-Options、X-Content-Type-Options 等
4. **HTTPS**：必须使用 HTTPS（PCI-DSS 4.1 要求）

完整的 NGINX 配置参考 `nginx/nginx.conf` 文件中的 User Frontend 部分。

---

## Admin 前端部署

Admin 前端同样是 Vue 3 SPA 应用，构建和部署方法与 User 前端完全相同。

### 构建步骤

```bash
cd admin

# 安装依赖
npm ci --ignore-scripts

# 配置 API 地址（可选，同 User 说明）
# export VITE_API_BASE_URL=https://api.your-domain.com

# 构建
npm run build

# 部署
sudo mkdir -p /var/www/admin/dist
sudo cp -r dist/* /var/www/admin/dist/

# 设置安全文件权限
sudo chown -R www-data:www-data /var/www/admin/dist
sudo find /var/www/admin/dist -type d -exec chmod 755 {} \;
sudo find /var/www/admin/dist -type f -exec chmod 644 {} \;
```

### 管理后台安全加固（重要！）

管理后台包含敏感操作（用户管理、订单管理、支付配置等），**必须**额外加固：

#### 1. 使用独立端口或子域名

推荐方案（任选其一）：

| 方案 | 说明 | 示例 |
|------|------|------|
| 独立端口 | 管理后台监听独立端口 | `https://your-domain.com:8443` |
| 子域名 | 管理后台使用独立子域名 | `https://admin.your-domain.com` |
| 路径前缀 | 管理后台在主域名下的子路径 | `https://your-domain.com/admin-panel/` |

#### 2. IP 白名单（强烈推荐）

在 NGINX 管理后台 server block 中限制 IP：

```nginx
server {
    # ...管理后台配置...

    # ⚠️ 仅允许管理员 IP 访问
    allow 你的办公IP/32;         # 你的固定公网 IP
    allow 10.0.0.0/8;             # 内网（如果需要）
    allow 192.168.0.0/16;         # 内网（如果需要）
    deny all;                     # 拒绝其他所有 IP
}
```

#### 3. 更严格的 CSP

管理后台不需要 Telegram OAuth 等外部脚本：

```nginx
add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob: https:; font-src 'self' data:; connect-src 'self'; object-src 'none'; base-uri 'self'; form-action 'self';" always;
```

#### 4. 禁止 iframe 嵌入

```nginx
add_header X-Frame-Options "DENY" always;
```

#### 5. 客户端证书认证（可选高安全方案）

对于高安全要求的环境，可配置 NGINX mTLS：

```nginx
server {
    # ...管理后台配置...
    ssl_client_certificate /etc/nginx/ssl/admin-ca.crt;
    ssl_verify_client on;
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

项目提供了针对 Dujiao-Next **整站**的 OWASP CRS 追加规则文件：
`nginx/modsecurity/dujiao-crs-rules.conf`

该文件覆盖了 **用户前台、管理后台、API 后端** 全部组件的安全配置。

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

### 规则总览

规则按组件分为四部分：

#### 第一部分: API 接口豁免规则 (10001-10020)

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

#### 第二部分: 用户前台安全规则 (10050-10069)

| 规则 ID | 说明 | 作用 |
|---------|------|------|
| 10050 | SPA HTML 响应豁免 | 防止 index.html 触发出站 XSS 误报 |
| 10051 | 静态资源豁免 | JS/CSS/字体/图片不触发 WAF |
| 10052 | 浏览器标准文件豁免 | favicon/robots/manifest |
| 10053 | 上传目录文件类型限制 | 仅允许图片文件类型 |

#### 第三部分: 管理后台安全规则 (10070-10089)

| 规则 ID | 说明 | 作用 |
|---------|------|------|
| 10071 | 操作日志豁免 | 管理员查看日志不触发误报 |
| 10072 | 邮件模板编辑豁免 | HTML 模板内容不触发 XSS 规则 |

#### 第四部分: 全站通用安全加固 (10100-10119)

| 规则 ID | 说明 | 作用 |
|---------|------|------|
| 10100 | 敏感文件屏蔽 | 阻止 .git/.env/.sql/.db 等文件访问 |
| 10101 | 隐藏文件屏蔽 | 阻止以 `.` 开头的文件/目录 |
| 10102 | 源码目录屏蔽 | 阻止 node_modules/vendor/internal 等 |
| 10103 | 配置文件屏蔽 | 阻止 config.yml/docker-compose 等 |
| 10104 | HTTP 方法限制 | 仅允许 GET/POST/PUT/PATCH/DELETE/OPTIONS |
| 10105 | PHP/ASP 探测阻止 | 阻止扫描器探测其他语言脚本 |
| 10106 | 后门路径阻止 | 阻止 wp-admin/phpMyAdmin 等 |
| 10107 | API 响应类型检查 | API 响应必须是 JSON |
| 10108 | 空 User-Agent 阻止 | 阻止无 UA 的异常请求 |

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
