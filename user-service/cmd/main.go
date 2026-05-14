package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"elibrary/gen/userpb"
	"elibrary/pkg/eventbus"
	"elibrary/user-service/internal/clients"
	grpcdelivery "elibrary/user-service/internal/delivery/grpc"
	"elibrary/user-service/internal/mailer"
	"elibrary/user-service/internal/repository/postgres"
	"elibrary/user-service/internal/usecase"
)

func main() {
	dsn := getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5433/elibrary_users?sslmode=disable")
	addr := getenv("GRPC_ADDR", ":50051")
	jwtSecret := getenv("JWT_SECRET", "change-me-in-production")
	natsURL := getenv("NATS_URL", "nats://localhost:4222")
	smtpHost := getenv("SMTP_HOST", "localhost")
	smtpPortStr := getenv("SMTP_PORT", "1025")
	smtpUser := getenv("SMTP_USER", "")
	smtpPass := getenv("SMTP_PASS", "")
	mailFrom := getenv("MAIL_FROM", "noreply@elibrary.local")
	verifyURL := getenv("VERIFY_URL", "http://localhost:8080/api/auth/verify?token=")
	borrowAddr := getenv("BORROW_GRPC_ADDR", "localhost:50053")

	smtpPort, _ := strconv.Atoi(smtpPortStr)

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

	pub, err := eventbus.NewPublisher(natsURL)
	if err != nil {
		log.Printf("nats unavailable: %v (continuing without events)", err)
	} else {
		defer pub.Close()
		log.Println("connected to nats")
	}

	m := mailer.New(smtpHost, smtpPort, smtpUser, smtpPass, mailFrom)

	bc, err := clients.NewBorrowClient(borrowAddr)
	if err != nil {
		log.Printf("borrow client unavailable: %v (continuing without it)", err)
	} else {
		defer bc.Close()
		log.Println("connected to borrow-service")
	}

	repo := postgres.NewUserRepo(pool)
	uc := usecase.NewUserUsecase(repo, jwtSecret, pub, m, bc, verifyURL)
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
