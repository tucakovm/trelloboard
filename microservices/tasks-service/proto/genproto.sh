
protoc \
  --proto_path=. \                          # Path to your proto files
  --proto_path=./googleapis \               # Path to Google's proto files
  --go_out=./task \                         # Generate Go code for gRPC (TaskService)
  --go_opt=paths=source_relative \          # Ensure relative imports in generated code
  --go-grpc_out=./task \                    # Generate Go server interface
  --go-grpc_opt=paths=source_relative \     # Ensure relative imports in generated gRPC code
  --grpc-gateway_out=./task \               # Generate gRPC Gateway code
  --grpc-gateway_opt=paths=source_relative \  # Ensure relative imports for gateway
  task.proto                                # The proto file to process
