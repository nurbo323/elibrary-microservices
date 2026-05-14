FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /tmp/api-gateway ./api-gateway/cmd

FROM alpine:latest

WORKDIR /app

COPY --from=builder /tmp/api-gateway .

EXPOSE 8080

CMD ["./api-gateway"]