FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /tmp/borrow-service ./borrow-service/cmd

FROM alpine:latest

WORKDIR /app

COPY --from=builder /tmp/borrow-service .

EXPOSE 50053

CMD ["./borrow-service"]