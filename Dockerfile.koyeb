# Use the official Golang image as a build stage
FROM golang:1.23-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN go build -o main .

RUN ls -la

# Use a minimal image for the final stage
FROM alpine:latest

# Set the working directory in the final image
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main .

# Dockerfile
COPY .env /app/.env

# Expose the port the app runs on (adjust if needed)
EXPOSE 8080

# Run the application
CMD ["./main"]