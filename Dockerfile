# Dockerfile
FROM golang:1.24.5 AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the server
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./server/main.go

# Final stage - sử dụng image có sẵn Chrome headless
FROM chromedp/headless-shell:latest

# Install ca-certificates for HTTPS requests
RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server /app/server

# Expose port
EXPOSE 8080

# Run the server
CMD ["/app/server"]