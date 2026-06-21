FROM golang:alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -ldflags="-s -w" -o /app/main ./cmd/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/.env /app/
COPY --from=builder /app/config ./config

EXPOSE 44046

ENTRYPOINT ["/app/main"]
CMD ["--config=/app/config/dev.yaml"]
