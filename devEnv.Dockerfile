FROM golang:1.23-alpine
RUN mkdir -p /app/src && \
    go env -w GOPROXY=https://goproxy.cn,direct && \
    go env -w CGO_ENABLED=0 && \
    go env -w GO111MODULE=on && \
    sed -i 's#https\?://dl-cdn.alpinelinux.org/alpine#https://mirrors.cernet.edu.cn/alpine#g' /etc/apk/repositories && \
    apk update && \
    apk add upx
VOLUME ./:/app
WORKDIR /app/src