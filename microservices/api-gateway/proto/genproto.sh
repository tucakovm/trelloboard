protoc \
  --proto_path=. \
  --proto_path=./googleapis \
  --go_out=./gateway \
  --go_opt=paths=source_relative \
  --go-grpc_out=./gateway \
  --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=./gateway \
  --grpc-gateway_opt=paths=source_relative \
  ./api-gateway.proto  || echo "Failed to generate code"
