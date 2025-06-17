# Multi-stage build for smaller image size (important for Pi)
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bot ./cmd/bot

# Final stage - minimal image for Pi
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/bot .

# Create non-root user
RUN addgroup -g 1000 -S dndbot && \
    adduser -u 1000 -S dndbot -G dndbot

USER dndbot

EXPOSE 8080

CMD ["./bot"]