# Этап сборки
FROM golang:1.22-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Этап запуска тестов
FROM golang:1.22-bookworm AS tester
WORKDIR /app
COPY --from=builder /app .
CMD ["go", "test", "./..."]

# Этап продакшен
FROM scratch AS production
COPY --from=builder /app/main /
EXPOSE 8080
CMD ["/main"]
