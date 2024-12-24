protoc \
  --proto_path=. \
  --proto_path=./googleapis \
  --go_out=./task \
  --go_opt=paths=source_relative \
  --go-grpc_out=./task \
  --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=./task \
  --grpc-gateway_opt=paths=source_relative \
  ./task.proto> protoc_output.log 2>&1
