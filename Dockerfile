# Builder
FROM golang:latest AS builder

WORKDIR /app
# install make util
RUN apt-get update && apt-get install -y make build-essential && rm -rf /var/lib/apt/lists/*
COPY . .
RUN make release

# Deploy
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/dist/exchange .
COPY --from=builder /app/app/exg.db .

# setup logs dir as volume, for mount
VOLUME ["/app/logs"]

EXPOSE 8080 8081

CMD ["./exchange"]