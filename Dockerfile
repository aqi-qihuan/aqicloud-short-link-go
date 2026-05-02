# Multi-stage build for all services
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/gateway ./cmd/gateway/
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/account ./cmd/account/
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/link ./cmd/link/
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/data ./cmd/data/
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/shop ./cmd/shop/
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/ai ./cmd/ai/

# --- Gateway ---
FROM alpine:3.19 AS gateway
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/gateway /usr/local/bin/
EXPOSE 8888
CMD ["gateway"]

# --- Account Service ---
FROM alpine:3.19 AS account
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/account /usr/local/bin/
EXPOSE 8001
CMD ["account"]

# --- Link Service ---
FROM alpine:3.19 AS link
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/link /usr/local/bin/
EXPOSE 8003
CMD ["link"]

# --- Data Service ---
FROM alpine:3.19 AS data
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/data /usr/local/bin/
EXPOSE 8002
CMD ["data"]

# --- Shop Service ---
FROM alpine:3.19 AS shop
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/shop /usr/local/bin/
EXPOSE 8005
CMD ["shop"]

# --- AI Service ---
FROM alpine:3.19 AS ai
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/ai /usr/local/bin/
EXPOSE 8006
CMD ["ai"]
