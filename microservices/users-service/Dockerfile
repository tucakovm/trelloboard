# Stage 1: Build the Go application
FROM golang:alpine as build_container
WORKDIR /app
COPY go.mod . 
RUN go mod download
COPY . . 
RUN go build -o server

# Stage 2: Final image with Redis and Go app
FROM alpine
WORKDIR /app

# Install Redis
RUN apk add --no-cache redis

# Copy the Go binary from the build stage
COPY --from=build_container /app/server /usr/bin/server

# Expose ports for the Go app and Redis
EXPOSE 8001 6379

# Command to start both Redis and the Go application
ENTRYPOINT ["sh", "-c", "redis-server --daemonize yes && server"]
