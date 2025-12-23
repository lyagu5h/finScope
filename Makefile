grpc_ledger:
	protoc -I ./proto \
	--go_out=./ledger --go_opt=paths=source_relative \
	--go-grpc_out=./ledger --go-grpc_opt=paths=source_relative \
	proto/internal/delivery/protos/ledger/v1/ledger.proto
	
grpc_gateway:
	protoc -I ./proto \
	--go_out=./gateway --go_opt=paths=source_relative \
	--go-grpc_out=./gateway --go-grpc_opt=paths=source_relative \
	proto/internal/delivery/protos/ledger/v1/ledger.proto
