# --- build stage ---
FROM golang:1.26-alpine AS builder

WORKDIR /app

# код и зависимости
COPY . .
RUN go mod download

# сборка бинарника
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd/app

# --- runtime stage ---
FROM alpine:3.19

WORKDIR /app

# копируем бинарник из builder
COPY --from=builder /app/app .

# копируем статические файлы
COPY --from=builder /app/assets ./assets

# создаем непривилегированного пользователя
RUN adduser -D -u 1000 appuser
USER appuser

# запуск
CMD ["./app"]