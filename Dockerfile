# 构建阶段
FROM golang:1.24 AS builder

WORKDIR /app

#ENV GOPROXY="https://goproxy.cn,direct"

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码并编译
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app .

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 从构建阶段复制编译好的程序
COPY --from=builder /app/app .

# 创建必要的目录
RUN mkdir -p db utils/OSS

# 复制配置文件（保持相对路径）
COPY db/.env db/
COPY utils/OSS/.env utils/OSS/
COPY config/.admin config/

EXPOSE 8080

CMD ["./app"]
