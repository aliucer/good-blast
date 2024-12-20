# Build stage
ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm AS builder

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o /run-app .

# Final stage
FROM debian:bookworm

# Install CA certificates, Redis, cron, curl
RUN apt-get update && apt-get install -y ca-certificates redis-server cron curl && rm -rf /var/lib/apt/lists/*

# Create /data directory for Redis persistence
RUN mkdir -p /data

COPY --from=builder /run-app /usr/local/bin/run-app

ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt

# Copy the rotate_tournament.sh script and make it executable
COPY rotate_tournament.sh /rotate_tournament.sh
RUN chmod +x /rotate_tournament.sh

# Copy crontab file
COPY crontab /etc/cron.d/rotate_tournament
RUN chmod 0644 /etc/cron.d/rotate_tournament
RUN crontab /etc/cron.d/rotate_tournament

# Create a start script to run Redis, cron, and the Go app
COPY start.sh /start.sh
RUN chmod +x /start.sh

# Expose the ports for the app (8080) and Redis (6379)
EXPOSE 8080 6379

CMD ["/start.sh"]
