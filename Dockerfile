FROM golang:1.22.1-alpine3.19 AS builder

COPY . .
WORKDIR /go

RUN pwd
RUN go mod download
RUN go build -o bin/chat_server cmd/main.go

FROM alpine:latest

WORKDIR /root/
COPY --from=builder /go/bin/chat_server .

CMD ["sh", "-c", "sleep 20 && ./chat_server"]