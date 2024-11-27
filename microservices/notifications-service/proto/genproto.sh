protoc \
  --proto_path=. \
  --proto_path=./googleapis \
  --go_out=./notification \
  --go_opt=paths=source_relative \
  --go-grpc_out=./notification \
  --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=./notification \
  --grpc-gateway_opt=paths=source_relative \
  ./notification.proto
