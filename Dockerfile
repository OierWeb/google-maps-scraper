# Build stage - Optimized for Browserless (no local browser needed)
FROM golang:1.24.0-bullseye AS builder
WORKDIR /app

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build the application
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /usr/bin/google-maps-scraper

# Final stage - Minimal runtime for Browserless connection
FROM debian:bullseye-slim

# Set environment variables for Browserless mode
ENV PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1
ENV PLAYWRIGHT_BROWSERS_PATH=0

# Install only essential runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Copy the compiled binary from builder stage
COPY --from=builder /usr/bin/google-maps-scraper /usr/bin/

# Create non-root user and necessary directories
RUN useradd -r -s /bin/false scraper \
    && mkdir -p /app/webdata /app/cache /app/results \
    && chown -R scraper:scraper /app

# Set working directory
WORKDIR /app

# Switch to non-root user
USER scraper

ENTRYPOINT ["google-maps-scraper"]
