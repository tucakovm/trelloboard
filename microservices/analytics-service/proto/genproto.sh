protoc \
  --proto_path=. \
  --proto_path=./googleapis \
  --go_out=./analytics \
  --go_opt=paths=source_relative \
  --go-grpc_out=./analytics \
  --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=./analytics \
  --grpc-gateway_opt=paths=source_relative \
  ./analytics.proto > protoc.log 2>&1 || echo "Failed to generate code. Check protoc.log for details."
