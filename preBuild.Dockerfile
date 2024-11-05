FROM golang:1.23-alpine as builder

RUN mkdir -p /app/src && \
    mkdir /dist
COPY ./ /app/src
WORKDIR /app/src
RUN mkdir "vendor" && \
    go mod tidy &&  \
    \ go mod vendor && \
    sed -i 's#https\?://dl-cdn.alpinelinux.org/alpine#https://mirrors.cernet.edu.cn/alpine#g' /etc/apk/repositories && \
    apk update && \
    apk add upx
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux
RUN go build -o /bin/mcdownload -ldflags="-s -w" script/download-go/main.go && \
    upx -9 /app/bin/mcdownload
RUN go build -o /app/bin/mclaunch -ldflags="-s -w" script/launch-go/main.go && \
    upx -9 /app/bin/mclaunch