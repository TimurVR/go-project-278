# 1) Install frontend package
FROM node:24-alpine AS frontend-installer
WORKDIR /build

# Устанавливаем только пакет фронтенда
RUN npm install @hexlet/project-url-shortener-frontend --no-audit --no-fund --prefer-offline

# 2) Build backend
FROM golang:1.25-alpine AS backend-builder
RUN apk add --no-cache git
WORKDIR /build/code

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /build/app .

# 3) Runtime
FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata bash caddy

WORKDIR /app

COPY --from=backend-builder /build/app /app/bin/app

# Копируем готовые статические файлы из установленного npm пакета
COPY --from=frontend-installer /build/node_modules/@hexlet/project-url-shortener-frontend/dist /app/public

COPY --from=backend-builder /build/code/db/migrations /app/db/migrations
COPY --from=backend-builder /go/bin/goose /usr/local/bin/goose

COPY bin/run.sh /app/bin/run.sh
RUN chmod +x /app/bin/run.sh

COPY Caddyfile /etc/Caddyfile

EXPOSE 80

CMD ["/app/bin/run.sh"]