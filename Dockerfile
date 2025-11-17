# 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /build

# 安装必要的工具
RUN apk add --no-cache git make

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译应用
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/go-uni-pay cmd/server/main.go

# 运行阶段
FROM alpine:latest

# 设置时区
RUN apk --no-cache add ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 创建应用目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/go-uni-pay /app/

# 复制配置文件
COPY configs /app/configs

# 创建日志目录
RUN mkdir -p /app/logs

# 暴露端口
EXPOSE 8080

# 设置非 root 用户
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app

USER appuser

# 启动应用
CMD ["/app/go-uni-pay"]
