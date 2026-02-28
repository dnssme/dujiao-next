# syntax=docker/dockerfile:1

# ── Stage 1: Build ──────────────────────────────────────────
FROM golang:1.25.3-alpine AS builder

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
RUN echo "Building for $TARGETOS/$TARGETARCH$TARGETVARIANT"

WORKDIR /src

ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN set -eux; \
    export GOOS="$TARGETOS" GOARCH="$TARGETARCH"; \
    if [ "$TARGETARCH" = "arm" ] && [ -n "$TARGETVARIANT" ]; then export GOARM="${TARGETVARIANT#v}"; fi; \
    if [ "$TARGETARCH" = "amd64" ] && [ -n "$TARGETVARIANT" ]; then export GOAMD64="${TARGETVARIANT#v}"; fi; \
    go build -trimpath -ldflags="-s -w" -o /out/dujiao-api ./cmd/server

# ── Stage 2: Runtime (CIS / PCI-DSS hardened) ───────────────
FROM alpine:3.21

# CIS 4.1 - 使用受信任的最小化基础镜像 + 仅安装运行时必需包
RUN apk --no-cache add ca-certificates tzdata wget \
    && rm -rf /var/cache/apk/*

# CIS 4.2 / PCI-DSS 8.6 - 创建非 root 用户运行应用
RUN addgroup -S appgroup && adduser -S appuser -G appgroup -h /app -s /sbin/nologin

WORKDIR /app

# 创建数据目录并设置权限（CIS 4.6 - 最小文件权限）
RUN mkdir -p /app/db /app/uploads /app/logs \
    && chown -R appuser:appgroup /app \
    && chmod 750 /app /app/db /app/uploads /app/logs

COPY --from=builder --chown=appuser:appgroup /out/dujiao-api /app/dujiao-api
COPY --chown=appuser:appgroup config.yml.example /app/config.yml.example

# CIS 4.6 - 确保二进制文件不可写
RUN chmod 550 /app/dujiao-api

# CIS 4.2 - 切换到非 root 用户
USER appuser:appgroup

EXPOSE 8080

# CIS 6.1 - 添加健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO /dev/null http://127.0.0.1:8080/health || exit 1

CMD ["./dujiao-api"]
