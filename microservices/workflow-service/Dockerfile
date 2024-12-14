
FROM golang:alpine as build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o workflow-service .

# Stage 2
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=build /app/workflow-service .

EXPOSE 8005

CMD ["./workflow-service"]
