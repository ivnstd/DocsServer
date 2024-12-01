FROM golang:1.22.5-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o docs-server ./cmd/docs-server/main.go

FROM alpine:latest
WORKDIR /root/

COPY --from=builder /app/docs-server /docs-server
COPY --from=builder /app/internal/configs /root/configs
COPY --from=builder /app/.env /root/.env

CMD ["/docs-server"]