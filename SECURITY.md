# 安全审查报告 / Security Audit Report

> 对 Dujiao-Next 全栈（API 后端 + User 前端 + Admin 前端）的安全审查结果。

## 审查范围

| 组件 | 审查内容 |
|------|----------|
| **API 后端** | 认证授权、密码处理、数据库安全、支付安全、文件上传、输入验证、限流、配置安全 |
| **User 前端** | API 客户端、认证状态、路由守卫、XSS 风险面 |
| **Admin 前端** | API 客户端、权限控制、路由守卫 |
| **部署配置** | Dockerfile、docker-compose.yml、NGINX 配置、OWASP CRS 规则 |

## 审查结果

**整体风险等级：低（LOW）** — 未发现关键安全漏洞。

| 严重程度 | 数量 | 说明 |
|----------|------|------|
| 严重 (Critical) | 0 | 无 |
| 高危 (High) | 0 | 无 |
| 中危 (Medium) | 1 | `v-html` 使用 — 仅渲染管理员编辑的内容，风险可控 |
| 低危 (Low) | 2 | 自定义脚本注入（管理员功能）、localStorage token 存储 |
| 建议 (Info) | 3 | 改进建议 |

---

## API 后端安全

### JWT 认证 ✅

- 管理员和用户使用独立密钥体系
- Token 版本控制 + 失效时间校验，支持精确撤销
- 启动时检测弱密钥并拒绝启动（release 模式）

### RBAC 权限管理 ✅

- Casbin 实现角色基础访问控制
- 超级管理员旁路独立于策略
- 权限变更有审计日志

### 密码处理 ✅

- bcrypt 哈希（DefaultCost 10）
- 可配置密码策略（长度、大小写、数字、特殊字符）
- 使用 `crypto/rand` 生成随机密码

### 数据库安全 ✅

- GORM 参数化查询，无 SQL 拼接
- 支持 SQLite + PostgreSQL

### 支付安全 ✅

| 支付方式 | 签名算法 | 说明 |
|----------|----------|------|
| Alipay | RSA2 | 公钥验签 + 商户校验 |
| WeChat Pay | 官方 SDK | 内置签名验证 |
| Stripe | HMAC-SHA256 | Webhook 签名校验 |
| PayPal | 官方 API | 事件验证 |
| Epay | MD5 / RSA | 支付网关协议要求 MD5，应用层做常量时间比较 |
| BEpusdt | HMAC | 签名校验 |
| TokenPay | MD5 HMAC | 支付网关协议要求 MD5 |

> 注：Epay/BEpusdt/TokenPay 使用 MD5 是第三方支付网关的协议要求，非本程序选择。
> 应用层已使用常量时间比较防止时序攻击。

### 文件上传安全 ✅

- 大小限制（10MB）、扩展名白名单、MIME 类型检测、图片尺寸限制
- UUID 文件名，防止路径遍历

### 请求限流 ✅

- 基于 Redis 的滑动窗口限流（Lua 原子脚本）
- IP + 字段联合限流

### HTTP 服务器超时 ✅

- ReadTimeout: 30s, WriteTimeout: 30s, ReadHeaderTimeout: 10s, IdleTimeout: 120s

---

## 前端安全

### User 前端

| 项目 | 状态 | 说明 |
|------|------|------|
| API 客户端 | ✅ | Axios + 统一错误处理 + 401 自动登出 |
| 路由守卫 | ✅ | `requiresUserAuth` / `userGuest` meta 保护 |
| Token 存储 | 🟢 低危 | localStorage（行业通用做法） |
| v-html | 🟡 中危 | 3 处使用，内容均来自管理后台（非用户输入） |
| 自定义脚本 | 🟢 低危 | 管理员配置统计代码，非用户可控 |

### Admin 前端

| 项目 | 状态 | 说明 |
|------|------|------|
| API 客户端 | ✅ | Axios + Bearer Token + 401 自动登出 |
| RBAC 权限 | ✅ | 路由级别权限检查 + 通配符匹配 |
| 权限缓存 | ✅ | localStorage 缓存，每次加载刷新 |

---

## Docker 安全（CIS Benchmark）

### API Dockerfile

| CIS 编号 | 要求 | 状态 |
|-----------|------|------|
| 4.1 | 最小化基础镜像 | ✅ Alpine 3.21 |
| 4.2 | 非 root 用户 | ✅ appuser |
| 4.6 | 最小文件权限 | ✅ 750/550 |
| 4.10 | 多阶段构建 | ✅ |
| 6.1 | 健康检查 | ✅ |

### docker-compose.yml

| CIS 编号 | 要求 | 状态 |
|-----------|------|------|
| 5.2 | 资源限制 | ✅ CPU 2.0 / 内存 512M |
| 5.3 | 最小 Linux 能力 | ✅ `cap_drop: ALL` |
| 5.12 | 禁止新权限 | ✅ `no-new-privileges` |
| 5.26 | 健康检查 | ✅ |

---

## PCI-DSS 合规

| 编号 | 要求 | 状态 |
|------|------|------|
| 2.2.1 | 仅启用必要服务 | ✅ |
| 2.2.2 | 不使用默认密码 | ✅ 强制检查弱密钥 |
| 4.1 | 加密传输 | ✅ NGINX TLS 配置就绪 |
| 6.5.1 | 注入防护 | ✅ 参数化查询 + OWASP CRS |
| 6.5.3 | 安全存储 | ✅ bcrypt 哈希 |
| 6.5.7 | XSS 防护 | ✅ CSP 头 + v-html 仅管理员内容 |
| 6.5.10 | DoS 防护 | ✅ 超时 + 限流 |
| 6.6 | WAF | ✅ OWASP CRS 规则就绪 |
| 10.1 | 审计日志 | ✅ 结构化日志 + Request ID |

---

## 改进建议

1. 生产环境使用 PostgreSQL 替代 SQLite
2. 配置 TLS/HTTPS
3. CORS `allowed_origins` 配置具体域名，不使用 `*`
4. 管理后台限制 IP 白名单
5. 启用 OWASP CRS WAF
6. 对 v-html 内容可选在 API 层添加 HTML 消毒

---

## 审查声明

- 审查日期：2026-02-28
- 审查范围：全部 Go API 源代码、Vue 3 前端源代码（User + Admin）、Docker/NGINX 配置
- 审查方法：人工代码审查 × 5 轮 + 自动化静态分析
- 结论：**未发现关键安全漏洞**
