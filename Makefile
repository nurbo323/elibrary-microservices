.PHONY: up down proto migrate-users migrate-books migrate-borrows

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

proto:
	protoc \
	  --go_out=. --go_opt=paths=source_relative \
	  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	  proto/user.proto proto/book.proto proto/borrow.proto
	@mkdir -p gen/userpb gen/bookpb gen/borrowpb
	@mv proto/*.pb.go gen/userpb/ 2>/dev/null; true

# Удобные таргеты для миграций
migrate-users:
	migrate -path ./user-service/migrations \
	  -database "postgres://postgres:postgres@localhost:5433/elibrary_users?sslmode=disable" up

migrate-books:
	migrate -path ./book-service/migrations \
	  -database "postgres://postgres:postgres@localhost:5434/elibrary_books?sslmode=disable" up

migrate-borrows:
	migrate -path ./borrow-service/migrations \
	  -database "postgres://postgres:postgres@localhost:5435/elibrary_borrows?sslmode=disable" up