FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/analytics-service .

FROM alpine:3.17

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /bin/analytics-service .

EXPOSE 8005

CMD ["./analytics-service"]
