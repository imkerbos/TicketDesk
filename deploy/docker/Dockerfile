# 构建阶段
FROM golang:1.25 AS builder

WORKDIR /app

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/ticketdesk ./cmd/ticketdesk

# 运行阶段
FROM debian:bookworm-slim

WORKDIR /app

# 安装必要的运行时依赖
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

# 设置时区
ENV TZ=Asia/Shanghai

# 从构建阶段复制二进制文件
COPY --from=builder /app/bin/ticketdesk /app/ticketdesk
COPY --from=builder /app/configs /app/configs

# 创建日志目录
RUN mkdir -p /app/logs

# 暴露端口
EXPOSE 8080

# 运行
CMD ["/app/ticketdesk", "-config", "/app/configs/config.yaml"]
