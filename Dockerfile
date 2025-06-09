# Builder
FROM golang:latest AS builder

WORKDIR /app
# install make util
RUN apt-get update && apt-get install -y make build-essential && rm -rf /var/lib/apt/lists/*
COPY . .
RUN make release

# Deploy
FROM golang:latest
WORKDIR /app
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/dist/exchange .
COPY --from=builder /app/exg.db .

# create logs dir
RUN mkdir -p /app/logs
# setup logs dir as volume, for mount
VOLUME ["/app/logs"]

EXPOSE 8080

CMD ["./exchange"]