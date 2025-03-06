protoc \
  --proto_path=. \
  --proto_path=./googleapis \
  --go_out=./composer \
  --go_opt=paths=source_relative \
  --go-grpc_out=./composer \
  --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=./composer \
  --grpc-gateway_opt=paths=source_relative \
  ./composer.proto
