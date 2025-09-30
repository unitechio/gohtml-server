# ============================
# Stage 1: Build Go binary
# ============================
FROM golang:1.24.5 AS builder

WORKDIR /app

# Copy go.mod và go.sum trước để cache dependency
COPY go.mod go.sum ./
RUN go mod download

# Copy toàn bộ code
COPY . .

# Build binary Go (static, không cần CGO)
RUN CGO_ENABLED=0 go build -o server ./server/main.go

# ============================
# Stage 2: Runtime
# ============================
FROM chromedp/headless-shell:latest

WORKDIR /app

# Copy binary từ builder stage
COPY --from=builder /app/server /app/server

# Copy entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Port cho Go server
EXPOSE 8080
ENV PORT=8080

# Override entrypoint mặc định (vốn là headless-shell)
ENTRYPOINT ["/entrypoint.sh"]
