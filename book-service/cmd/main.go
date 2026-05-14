package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"elibrary/book-service/internal/cache"
	grpcdelivery "elibrary/book-service/internal/delivery/grpc"
	"elibrary/book-service/internal/repository/postgres"
	"elibrary/book-service/internal/usecase"
	"elibrary/pkg/eventbus"

	"elibrary/gen/bookpb"
)

func main() {
	dsn := getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5434/elibrary_books?sslmode=disable")
	addr := getenv("GRPC_ADDR", ":50052")
	natsURL := getenv("NATS_URL", "nats://localhost:4222")
	redisAddr := getenv("REDIS_ADDR", "localhost:6379")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	log.Println("connected to postgres")

	var c *cache.BookCache
	if bc, err := cache.New(redisAddr, 5*time.Minute); err != nil {
		log.Printf("redis unavailable: %v (continuing without cache)", err)
	} else {
		c = bc
		defer c.Close()
		log.Println("connected to redis")
	}

	pub, err := eventbus.NewPublisher(natsURL)
	if err != nil {
		log.Printf("nats unavailable: %v (continuing without events)", err)
	} else {
		defer pub.Close()
		log.Println("connected to nats")
	}

	bookRepo := postgres.NewBookRepo(pool)
	copyRepo := postgres.NewCopyRepo(pool)

	// Используем New (так как в твоем коде конструктор называется New)
	uc := usecase.New(bookRepo, copyRepo, c, pub)
	handler := grpcdelivery.NewBookHandler(uc)

	srv := grpc.NewServer()
	bookpb.RegisterBookServiceServer(srv, handler)
	reflection.Register(srv)

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
