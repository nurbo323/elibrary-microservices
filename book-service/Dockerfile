FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /tmp/book-service ./book-service/cmd

FROM alpine:latest

WORKDIR /app

COPY --from=builder /tmp/book-service .

EXPOSE 50052

CMD ["./book-service"]