FROM golang:alpine AS builder

RUN apk add --no-cache gcc musl-dev librdkafka-dev pkgconf

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -ldflags="-s -w" -o /app/main ./cmd/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates librdkafka
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/.env /app/
COPY --from=builder /app/config ./config

ENTRYPOINT ["/app/main"]
CMD ["--config=/app/config/dev.yaml"]
