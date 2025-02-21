# Build stage
FROM golang:1.23.6-alpine

# Install build dependencies for SQLite
RUN apk add --no-cache gcc musl-dev

# Enable CGO
ENV CGO_ENABLED=1

# Set working directory
WORKDIR /app

# Copy and download dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary with CGO enabled
RUN go build -o main .

# App port
EXPOSE 4000

# Run application
CMD ["./main"]