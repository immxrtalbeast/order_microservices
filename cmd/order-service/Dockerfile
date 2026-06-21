FROM golang:alpine AS builder

ENV GOPROXY=https://proxy.golang.org,direct
ENV GOWORK=off
ENV GOFLAGS=-mod=mod

WORKDIR /app
COPY internal/pkg/tracing/ ./internal/pkg/tracing/
COPY cmd/order-service/go.mod cmd/order-service/go.sum ./cmd/order-service/

WORKDIR /app/cmd/order-service
RUN --mount=type=cache,target=/go/pkg/mod go mod download

WORKDIR /app
COPY cmd/order-service/ ./cmd/order-service/

WORKDIR /app/cmd/order-service
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags="-s -w" -o /main ./cmd/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /main ./main
COPY cmd/order-service/.env /app/
COPY cmd/order-service/config /app/config

EXPOSE 44046
ENTRYPOINT ["/app/main"]
CMD ["--config=/app/config/dev.yaml"]
