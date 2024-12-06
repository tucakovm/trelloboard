protoc \
  --proto_path=. \
  --proto_path=./googleapis \
  --go_out=./workflows \
  --go_opt=paths=source_relative \
  --go-grpc_out=./workflows \
  --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=./workflows \
  --grpc-gateway_opt=paths=source_relative \
  ./workflows.proto > protoc_output.log 2>&1
