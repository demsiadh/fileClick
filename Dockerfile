# 第一阶段：构建Go后端
FROM golang:1.24 AS builder

WORKDIR /app

# 复制go.mod和go.sum（如果存在）
COPY go.mod ./

# 复制其余后端代码
COPY . .

# 构建Go二进制文件
RUN CGO_ENABLED=0 GOOS=linux go build -o fileClick

# 第二阶段：运行应用
FROM alpine:latest AS runtime

# 从构建阶段复制构建好的Go二进制文件
COPY --from=builder /app/fileClick /usr/local/bin/fileClick

# 复制数据目录
COPY data /data

# 暴露后端端口
EXPOSE 8080

# 启动后端
CMD fileClick -addr :8080