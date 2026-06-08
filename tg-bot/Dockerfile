FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY tg-bot/go.mod tg-bot/go.sum ./
RUN go mod download

COPY tg-bot/. .
RUN CGO_ENABLED=0 GOOS=linux go build -o arbiter .

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/arbiter /app/arbiter

CMD ["/app/arbiter"]
