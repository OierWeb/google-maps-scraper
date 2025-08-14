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

# Install essential runtime dependencies including libraries needed by Playwright
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libglib2.0-0 \
    libnss3 \
    libnspr4 \
    libatk1.0-0 \
    libatk-bridge2.0-0 \
    libcups2 \
    libdrm2 \
    libdbus-1-3 \
    libxkbcommon0 \
    libx11-6 \
    libxcomposite1 \
    libxdamage1 \
    libxext6 \
    libxfixes3 \
    libxrandr2 \
    libgbm1 \
    libpango-1.0-0 \
    libcairo2 \
    libasound2 \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Copy the compiled binary from builder stage
COPY --from=builder /usr/bin/google-maps-scraper /usr/bin/

# Create a startup script to ensure proper environment configuration
RUN echo '#!/bin/sh\n\
echo "🔧 Setting up Browserless environment"\n\
# Ensure PLAYWRIGHT_WS_ENDPOINT is set from BROWSER_WS_ENDPOINT if needed\n\
if [ -n "$BROWSER_WS_ENDPOINT" ] && [ -z "$PLAYWRIGHT_WS_ENDPOINT" ]; then\n\
  export PLAYWRIGHT_WS_ENDPOINT="$BROWSER_WS_ENDPOINT"\n\
  echo "📡 Set PLAYWRIGHT_WS_ENDPOINT=$PLAYWRIGHT_WS_ENDPOINT"\n\
fi\n\
\n\
# Ensure we never download browsers locally\n\
export PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1\n\
export PLAYWRIGHT_BROWSERS_PATH="/tmp/empty-browsers-path"\n\
export PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS=1\n\
export PLAYWRIGHT_DRIVER_PATH=""\n\
export PLAYWRIGHT_SKIP_BROWSER_GC=1\n\
export PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH=""\n\
export PLAYWRIGHT_FIREFOX_EXECUTABLE_PATH=""\n\
export PLAYWRIGHT_WEBKIT_EXECUTABLE_PATH=""\n\
\n\
# Create empty browsers directory\n\
mkdir -p /tmp/empty-browsers-path\n\
\n\
# Execute the original command\n\
exec "$@"' > /usr/bin/docker-entrypoint.sh \
    && chmod +x /usr/bin/docker-entrypoint.sh

# Create non-root user and necessary directories
RUN useradd -r -s /bin/false scraper \
    && mkdir -p /app/webdata /app/cache /app/results /home/scraper /tmp/playwright \
    && chown -R scraper:scraper /app /home/scraper /tmp/playwright \
    && chmod 755 /home/scraper

# Set working directory
WORKDIR /app

# Switch to non-root user
USER scraper

# Use our entrypoint script to ensure proper environment setup
ENTRYPOINT ["/usr/bin/docker-entrypoint.sh", "google-maps-scraper"]
