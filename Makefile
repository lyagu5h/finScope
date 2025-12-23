ENV_FILE := .env

ifneq ("$(wildcard $(ENV_FILE))","")
	include $(ENV_FILE)
	export
endif

proto:
	protoc -I ./proto \
	--go_out=./ledger --go_opt=paths=source_relative \
	--go-grpc_out=./ledger --go-grpc_opt=paths=source_relative \
	proto/internal/delivery/protos/ledger/v1/ledger.proto
	
	protoc -I ./proto \
	--go_out=./gateway --go_opt=paths=source_relative \
	--go-grpc_out=./gateway --go-grpc_opt=paths=source_relative \
	proto/internal/delivery/protos/ledger/v1/ledger.proto

migrate-up:
   goose -dir ./ledger/migrations postgres "$(DATABASE_URL)" up

migrate-down:
   goose -dir ./ledger/migrations postgres "$(DATABASE_URL)" down


test:
	go test ./...


	