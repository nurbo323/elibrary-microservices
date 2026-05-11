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
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"elibrary/borrow-service/internal/client"
	grpcdelivery "elibrary/borrow-service/internal/delivery/grpc"
	"elibrary/borrow-service/internal/repository/postgres"
	"elibrary/borrow-service/internal/usecase"
	"elibrary/gen/bookpb"
	"elibrary/gen/borrowpb"
	"elibrary/gen/userpb"
)

func main() {
	dsn := getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5435/elibrary_borrows?sslmode=disable")
	addr := getenv("GRPC_ADDR", ":50053")
	userAddr := getenv("USER_GRPC_ADDR", "localhost:50051")
	bookAddr := getenv("BOOK_GRPC_ADDR", "localhost:50052")

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

	userConn := mustDial(userAddr, "user-service")
	bookConn := mustDial(bookAddr, "book-service")
	defer userConn.Close()
	defer bookConn.Close()

	repo := postgres.NewBorrowRepo(pool)
	userCli := client.NewUserClient(userpb.NewUserServiceClient(userConn))
	bookCli := client.NewBookClient(bookpb.NewBookServiceClient(bookConn))

	uc := usecase.NewBorrowUsecase(repo, userCli, bookCli)
	handler := grpcdelivery.NewBorrowHandler(uc)

	srv := grpc.NewServer()
	borrowpb.RegisterBorrowServiceServer(srv, handler)
	reflection.Register(srv)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	go func() {
		log.Printf("borrow-service listening on %s", addr)
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

func mustDial(addr, name string) *grpc.ClientConn {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial %s (%s): %v", name, addr, err)
	}
	log.Printf("connected to %s at %s", name, addr)
	return conn
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}