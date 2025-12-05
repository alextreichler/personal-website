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

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/personal-website .

# Copy web assets (templates, static files)
COPY web ./web
# Copy migrations if they are needed at runtime (though currently db.go handles basic ones)
COPY migrations ./migrations

# Create a directory for the SQLite database
RUN mkdir -p data

# Expose the port the app runs on
EXPOSE 8080

# Command to run the application
CMD ["./personal-website"]
