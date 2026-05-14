package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"elibrary/api-gateway/internal/handler"

	// ИСПРАВЛЕНО: Добавлен правильный путь к твоим сгенерированным файлам
	"elibrary/elibrary/gen/bookpb"
	"elibrary/elibrary/gen/borrowpb"
	"elibrary/elibrary/gen/userpb"
)

func main() {
	httpAddr := getenv("HTTP_ADDR", ":8080")
	userAddr := getenv("USER_GRPC_ADDR", "localhost:50051")
	bookAddr := getenv("BOOK_GRPC_ADDR", "localhost:50052")
	borrowAddr := getenv("BORROW_GRPC_ADDR", "localhost:50053")

	userConn := mustDial(userAddr, "user-service")
	bookConn := mustDial(bookAddr, "book-service")
	borrowConn := mustDial(borrowAddr, "borrow-service")

	defer userConn.Close()
	defer bookConn.Close()
	defer borrowConn.Close()

	userH := handler.NewUserHandler(userpb.NewUserServiceClient(userConn))
	bookH := handler.NewBookHandler(bookpb.NewBookServiceClient(bookConn))
	borrowH := handler.NewBorrowHandler(borrowpb.NewBorrowServiceClient(borrowConn))

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(15 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/api", func(r chi.Router) {
		userH.Register(r)
		bookH.Register(r)
		borrowH.Register(r)
	})

	srv := &http.Server{
		Addr:    httpAddr,
		Handler: r,
	}

	go func() {
		log.Printf("api-gateway listening on %s", httpAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	log.Println("api-gateway stopped")
}

func mustDial(addr, name string) *grpc.ClientConn {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("dial %s (%s): %v", name, addr, err)
	}

	log.Printf("connected to %s at %s", name, addr)
	return conn
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
