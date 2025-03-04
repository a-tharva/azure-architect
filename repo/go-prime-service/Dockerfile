FROM golang:1.20 AS builder
WORKDIR /app
COPY . /app
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o prime-service main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/prime-service .
EXPOSE 50051
CMD ["./prime-service"]