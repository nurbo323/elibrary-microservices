.PHONY: up down logs proto clean-proto migrate-users migrate-books migrate-borrows

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

clean-proto:
	@if exist gen\\userpb rmdir /s /q gen\\userpb
	@if exist gen\\bookpb rmdir /s /q gen\\bookpb
	@if exist gen\\borrowpb rmdir /s /q gen\\borrowpb
	@if exist proto\\*.pb.go del /q proto\\*.pb.go

proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/user.proto proto/book.proto proto/borrow.proto
	@if not exist gen\\userpb mkdir gen\\userpb
	@if not exist gen\\bookpb mkdir gen\\bookpb
	@if not exist gen\\borrowpb mkdir gen\\borrowpb
	move proto\\user.pb.go gen\\userpb\\
	move proto\\user_grpc.pb.go gen\\userpb\\
	move proto\\book.pb.go gen\\bookpb\\
	move proto\\book_grpc.pb.go gen\\bookpb\\
	move proto\\borrow.pb.go gen\\borrowpb\\
	move proto\\borrow_grpc.pb.go gen\\borrowpb\\

migrate-users:
	migrate -path ./user-service/migrations -database "postgres://postgres:postgres@localhost:5433/elibrary_users?sslmode=disable" up

migrate-books:
	migrate -path ./book-service/migrations -database "postgres://postgres:postgres@localhost:5434/elibrary_books?sslmode=disable" up

migrate-borrows:
	migrate -path ./borrow-service/migrations -database "postgres://postgres:postgres@localhost:5435/elibrary_borrows?sslmode=disable" up