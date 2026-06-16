FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN go build -ldflags="-w -s -extldflags '-static'" \
    -tags timetzdata \
    -o karez-server .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/karez-server .
COPY --from=builder /app/config ./config
COPY --from=builder /app/.env.example ./.env

EXPOSE 8080 6060

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./karez-server"]
