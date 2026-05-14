FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /tmp/user-service ./user-service/cmd

FROM alpine:latest

WORKDIR /app

COPY --from=builder /tmp/user-service .

EXPOSE 50051

CMD ["./user-service"]