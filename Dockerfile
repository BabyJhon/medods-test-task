FROM golang:latest AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin ./cmd/main.go

RUN wget -O /usr/local/bin/goose https://github.com/pressly/goose/releases/download/v3.14.0/goose_linux_x86_64 \
    && chmod +x /usr/local/bin/goose

FROM alpine:3.18

WORKDIR /app

RUN apk add --no-cache postgresql-client

COPY --from=builder /app/bin /app/bin
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /usr/local/bin/goose /usr/local/bin/goose
COPY entrypoint.sh /app/entrypoint.sh
COPY .env /app/.env

RUN chmod +x /app/entrypoint.sh

RUN ls -l /app

CMD ["/app/bin"]