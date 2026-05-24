# ---- Stage 1: Build ----
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Cache dependencies separately from source code
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 = fully static binary (no C deps)
# -trimpath = remove host paths from binary
# -ldflags="-s -w" = strip debug info, smaller image
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o server .

# ---- Stage 2: Run ----
FROM alpine:3.19

RUN apk add --no-cache ca-certificates

# Non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/server .

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

# SIGTERM is handled in main.go via signal.Notify (graceful shutdown 10s)
ENTRYPOINT ["/app/server"]
