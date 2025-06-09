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
COPY --from=builder /app/exg.db .

RUN adduser -D -s /bin/sh appuser
# create logs dir
RUN mkdir -p /app/logs && chown -R appuser:appuser /app
USER appuser
# setup logs dir as volume, for mount
VOLUME ["/app/logs"]

EXPOSE 8080

CMD ["./exchange"]