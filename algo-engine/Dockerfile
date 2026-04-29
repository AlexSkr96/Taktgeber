FROM golang:1.24-alpine AS builder

WORKDIR /src
COPY src/ ./

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/algo-engine ./main.go

FROM gcr.io/distroless/static-debian12

WORKDIR /app
COPY --from=builder /out/algo-engine /app/algo-engine

USER nonroot:nonroot
ENTRYPOINT ["/app/algo-engine"]
