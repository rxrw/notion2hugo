FROM golang:1.22-alpine

WORKDIR /app

# 安装必要的依赖
RUN apk add --no-cache git

# 复制源代码
COPY . .

# 构建应用
RUN go build -o /usr/local/bin/notion2md ./cmd/notion2md

# 创建工作目录
WORKDIR /workspace

# 设置入口点脚本
COPY docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

ENTRYPOINT ["docker-entrypoint.sh"]
