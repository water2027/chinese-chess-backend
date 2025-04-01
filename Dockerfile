# 使用 Golang 进行编译
FROM golang:1.24 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN export GOPROXY=https://goproxy.cn,direct && go mod download
COPY . .
RUN go build -o main .

COPY config.json .
CMD ["./main"]