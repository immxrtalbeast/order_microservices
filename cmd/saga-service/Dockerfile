FROM golang:alpine AS builder

ENV GOPROXY=https://proxy.golang.org,direct
ENV GOWORK=off
ENV GOFLAGS=-mod=mod

WORKDIR /app
COPY internal/pkg/tracing/ ./internal/pkg/tracing/
COPY cmd/saga-service/go.mod cmd/saga-service/go.sum ./cmd/saga-service/

WORKDIR /app/cmd/saga-service
RUN --mount=type=cache,target=/go/pkg/mod go mod download

WORKDIR /app
COPY cmd/saga-service/ ./cmd/saga-service/

WORKDIR /app/cmd/saga-service
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags="-s -w" -o /main ./cmd/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /main ./main
COPY cmd/saga-service/.env /app/
COPY cmd/saga-service/config /app/config

ENTRYPOINT ["/app/main"]
CMD ["--config=/app/config/dev.yaml"]
