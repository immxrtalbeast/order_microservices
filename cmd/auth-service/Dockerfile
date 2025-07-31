FROM golang:alpine AS builder

# sudo docker run -e CONFIG_PATH=/app/config/local.yaml -p 8000:8000  c0dys/plandstu-go
RUN apk add --no-cache
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download -x
COPY . .             

RUN go build -ldflags="-s -w" -o /app/main ./cmd/main.go

FROM alpine:latest
RUN apk add --no-cache 
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/.env /app/
COPY --from=builder /app/config ./config

EXPOSE 8080

ENTRYPOINT ["/app/main"]
CMD ["--config=/app/config/local.yaml"]  # Используем абсолютный путь