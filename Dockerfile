# ========================================
# 🏗️  BUILD STAGE - アプリケーションのビルド
# ========================================
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git and ca-certificates for go mod download
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# ========================================
# 🚀 FINAL STAGE - 実行用イメージ
# ========================================
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Expose port (if needed)
EXPOSE 8080

# Run the binary
CMD ["./main"]
