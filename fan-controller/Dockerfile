FROM golang:1.15.5-alpine AS builder

# 启用go module
ENV GO111MODULE=on \
  GOPROXY=https://goproxy.cn,direct

WORKDIR /app

COPY main.go go.mod go.sum ./
COPY front-end/build ./build

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -o ./app ./main.go

# target
FROM arm32v7/debian:stable-slim

WORKDIR /app
COPY --from=builder /app .

EXPOSE 10001

CMD ["./app"]

