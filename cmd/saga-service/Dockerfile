FROM golang:alpine AS builder

ENV GOPROXY=https://proxy.golang.org,direct
ENV GOWORK=off
ENV GOFLAGS=-mod=mod

WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags="-s -w" -o /app/main ./cmd/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/.env /app/
COPY --from=builder /app/config ./config

ENTRYPOINT ["/app/main"]
CMD ["--config=/app/config/dev.yaml"]
