# Dockerfile
FROM golang:1.24.5 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server ./server/main.go

# Sử dụng image có Chrome nhưng override CMD
FROM chromedp/headless-shell:latest

WORKDIR /app
COPY --from=builder /app/server /app/server

EXPOSE 8080
ENV PORT=8080

# QUAN TRỌNG: Override CMD để chạy server Go
CMD ["/app/server"]