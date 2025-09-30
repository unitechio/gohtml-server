FROM golang:1.24.5 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server ./server/main.go

FROM chromedp/headless-shell:latest
WORKDIR /app
COPY --from=builder /app/server /app/server

EXPOSE 8080
ENV PORT=8080

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

CMD ["/entrypoint.sh"]
