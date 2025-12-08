FROM golang:1.24.0-alpine AS builder

# Install build tools
RUN apk add --no-cache git

# Set work directory
WORKDIR /app

# Only copy go.mod & go.sum first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy entire source
COPY . .

# Build binary
RUN go build -o kiosk-backend ./src/main.go


# --------------------------------------
# 2. Runtime stage
# --------------------------------------
FROM alpine:3.20

# Add CA certs (needed for HTTPs calls)
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/kiosk-backend .

# Set permissions
RUN chmod +x kiosk-backend

# Set the port your server listens on
EXPOSE 17069

# Run the app
CMD ["./kiosk-backend"]
