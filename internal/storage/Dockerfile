# 基础镜像
FROM golang:1.20-alpine3.18 as builder

# 设置工作目录
WORKDIR /app

# 将本地文件复制到容器中
COPY . .

RUN apk update
RUN apk add make

# 下载依赖并编译应用程序
RUN make build

FROM alpine
COPY --from=builder /app/bin/main .
ENTRYPOINT ["/main"]