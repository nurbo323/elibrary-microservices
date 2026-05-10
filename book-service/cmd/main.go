package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	grpcdelivery "elibrary/book-service/internal/delivery/grpc"
	"elibrary/book-service/internal/repository/postgres"
	"elibrary/book-service/internal/usecase"
	"elibrary/gen/bookpb"
)

func main() {
	// Подключаемся к базе elibrary_books на порту 5434 [cite: 3464]
	dsn := getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5434/elibrary_books?sslmode=disable")
	addr := getenv("GRPC_ADDR", ":50052")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Настройка пула соединений с БД [cite: 3464]
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	log.Println("connected to postgres")

	// Инициализация слоев (Clean Architecture) [cite: 3464, 3482]
	bookRepo := postgres.NewBookRepo(pool)
	copyRepo := postgres.NewCopyRepo(pool)
	uc := usecase.NewBookUsecase(bookRepo, copyRepo)
	handler := grpcdelivery.NewBookHandler(uc)

	// Создание и настройка gRPC сервера [cite: 3464]
	srv := grpc.NewServer()
	bookpb.RegisterBookServiceServer(srv, handler)
	reflection.Register(srv) // Нужно для работы grpcurl [cite: 3464]

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	go func() {
		log.Printf("book-service listening on %s", addr)
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("serve: %v", err)
		}
	}()

	// Ожидание сигнала на остановку (Ctrl+C) [cite: 3464]
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("shutting down")
	srv.GracefulStop()
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
