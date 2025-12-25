# Build Stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy dependency definitions
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the application
# CGO_ENABLED=0 is used for a static build, which is ideal for 'scratch' or 'alpine' containers
RUN CGO_ENABLED=0 GOOS=linux go build -o personal-website ./cmd/server

# Run Stage
FROM alpine:latest

# Create a non-root user and group
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/personal-website .

# Copy web assets (templates, static files)
COPY web ./web

# Create necessary directories with correct permissions
RUN mkdir -p data web/static/uploads && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose the port the app runs on
EXPOSE 8080

# Command to run the application
CMD ["./personal-website"]
