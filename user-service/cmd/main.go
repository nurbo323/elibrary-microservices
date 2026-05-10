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

	"elibrary/gen/userpb"
	grpcdelivery "elibrary/user-service/internal/delivery/grpc"
	"elibrary/user-service/internal/repository/postgres"
	"elibrary/user-service/internal/usecase"
)

func main() {
	dsn := getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5433/elibrary_users?sslmode=disable")
	addr := getenv("GRPC_ADDR", ":50051")
	jwtSecret := getenv("JWT_SECRET", "change-me-in-production")

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

	repo := postgres.NewUserRepo(pool)
	uc := usecase.NewUserUsecase(repo, jwtSecret)
	handler := grpcdelivery.NewUserHandler(uc)

	srv := grpc.NewServer()
	userpb.RegisterUserServiceServer(srv, handler)
	reflection.Register(srv)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	go func() {
		log.Printf("user-service listening on %s", addr)
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
