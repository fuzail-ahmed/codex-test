.PHONY: generate migrate-up migrate-down run test

# PROTO=proto/todo/v1/todo.proto

PROTO_DIR := proto/todo/v1
PROTO_SRC := $(wildcard $(PROTO_DIR)/*.proto)
GO_OUT := shared/gen/todo/v1

# Generate gRPC stubs

.PHONY: generate-proto
generate-proto:
	protoc \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(GO_OUT) \
		--go-grpc_out=$(GO_OUT) \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		$(PROTO_SRC)

migrate-up:
	go run ./cmd/migrate -database "$(DB_DSN)" -command up

migrate-down:
	go run ./cmd/migrate -database "$(DB_DSN)" -command down

run:
	go run ./cmd/todo-api

test:
	go test ./...