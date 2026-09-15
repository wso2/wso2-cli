# Build stage
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod tidy && go get -u all
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o choreo ./cmd/main.go

# Runtime stage
FROM alpine:3.18
WORKDIR /app

# Create a non-root user and group with a UID/GID in the specified range
# The -D flag creates a system user without a home directory, suitable for containers.
RUN addgroup -g 10001 appgroup && adduser -u 10001 -G appgroup -D appuser

# Copy the compiled binary
COPY --from=builder /app/choreo /usr/local/bin/choreo

# Switch to the non-root user using its numeric ID
USER 10001

# Expose the port (default 8080)
EXPOSE 8080

# Allow PORT to be overridden at runtime
ENV PORT=8080

# Run the MCP server in HTTP mode, listening on the specified port
ENTRYPOINT ["/bin/sh", "-c", "/usr/local/bin/choreo start-mcp-server --http --port ${PORT}"]