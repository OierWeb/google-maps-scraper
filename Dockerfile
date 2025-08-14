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

# Set environment variables to FORCE remote browser usage only
ENV PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1
ENV PLAYWRIGHT_BROWSERS_PATH=""
ENV PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS=1
ENV PLAYWRIGHT_DRIVER_PATH=""
ENV HOME=/home/scraper
ENV PLAYWRIGHT_SKIP_BROWSER_GC=1
ENV PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH=""
ENV PLAYWRIGHT_FIREFOX_EXECUTABLE_PATH=""
ENV PLAYWRIGHT_WEBKIT_EXECUTABLE_PATH=""

# Install only essential runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Copy the compiled binary from builder stage
COPY --from=builder /usr/bin/google-maps-scraper /usr/bin/

# Create non-root user and necessary directories
RUN useradd -r -s /bin/false scraper \
    && mkdir -p /app/webdata /app/cache /app/results /home/scraper /tmp/playwright \
    && chown -R scraper:scraper /app /home/scraper /tmp/playwright \
    && chmod 755 /home/scraper

# Set working directory
WORKDIR /app

# Switch to non-root user
USER scraper

ENTRYPOINT ["google-maps-scraper"]
