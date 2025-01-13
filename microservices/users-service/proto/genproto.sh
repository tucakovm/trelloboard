protoc \
  --proto_path=. \
  --proto_path=./googleapis \
  --go_out=./users \
  --go_opt=paths=source_relative \
  --go-grpc_out=./users \
  --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=./users \
  --grpc-gateway_opt=paths=source_relative \
  ./users.proto
