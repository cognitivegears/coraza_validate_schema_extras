# Use a Go base image
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy module files first for caching
COPY go.mod ./
# Copy go.sum if it exists
COPY go.sum* ./

# Copy Coraza module files
COPY coraza/go.mod coraza/go.sum coraza/
# Copy Coraza workspace file
COPY coraza/go.work coraza/

# Copy the Coraza submodule source
# Ensure the submodule is initialized locally before building: git submodule update --init
COPY coraza/ /app/coraza/

# Download dependencies for the main module and the submodule
RUN go mod download || go mod tidy

# Copy the rest of the application source code
COPY . .

# Fix missing dependencies before building
RUN go get rsc.io/binaryregexp
RUN go mod tidy

# Build the application
# Use CGO_ENABLED=0 for static linking if needed, but Coraza might require CGO
# The -ldflags="-s -w" flags strip debugging information and symbols to reduce binary size.
RUN go build -ldflags="-s -w" -o /validate-server server.go

# --- Final Stage ---
FROM alpine:latest

# Install any necessary runtime dependencies if required (e.g., ca-certificates)
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /validate-server /app/validate-server

# Copy default rules and schemas into the image
# These can be overridden by the volume mount

# Define the default path for rules, which can be overridden by environment variable
ENV CORAZA_RULES_DIR=/app

# Expose the default port the server runs on
EXPOSE 8080

# Define the entry point
# The server will respect the CORAZA_RULES_DIR environment variable
ENTRYPOINT ["/app/validate-server"]

# Default command (can be overridden) - specify the port
CMD ["-port", "8080"]
