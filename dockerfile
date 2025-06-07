# Builder
FROM golang:latest AS builder
WORKDIR /app
COPY . .
RUN go build -o exchange

# Deploy
FROM golang:latest
WORKDIR /app
COPY --from=builder /app/exchange .
EXPOSE 8080
CMD ["./exchange"]