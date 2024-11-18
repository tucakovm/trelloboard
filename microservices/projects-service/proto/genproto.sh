protoc \
  --proto_path=. \
  --proto_path=./googleapis \
  --go_out=./project \
  --go_opt=paths=source_relative \
  --go-grpc_out=./project \
  --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=./project \
  --grpc-gateway_opt=paths=source_relative \
  ./project.proto
