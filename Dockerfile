# Build stage for Playwright dependencies (conditional)
FROM golang:1.24.3-bullseye AS playwright-deps
ARG USE_BROWSERLESS=false
ENV PLAYWRIGHT_BROWSERS_PATH=/opt/browsers
#ENV PLAYWRIGHT_DRIVER_PATH=/opt/
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    && curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && go install github.com/playwright-community/playwright-go/cmd/playwright@latest \
    && mkdir -p /opt/browsers \
    && if [ "$USE_BROWSERLESS" != "true" ]; then playwright install chromium --with-deps; fi

# Build stage
FROM golang:1.24.3-bullseye AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /usr/bin/google-maps-scraper

# Final stage
FROM debian:bullseye-slim
ARG USE_BROWSERLESS=false
ENV PLAYWRIGHT_BROWSERS_PATH=/opt/browsers
ENV PLAYWRIGHT_DRIVER_PATH=/opt

# Install dependencies conditionally based on Browserless usage
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    $(if [ "$USE_BROWSERLESS" != "true" ]; then echo "\
    libnss3 \
    libnspr4 \
    libatk1.0-0 \
    libatk-bridge2.0-0 \
    libcups2 \
    libdrm2 \
    libdbus-1-3 \
    libxkbcommon0 \
    libatspi2.0-0 \
    libx11-6 \
    libxcomposite1 \
    libxdamage1 \
    libxext6 \
    libxfixes3 \
    libxrandr2 \
    libgbm1 \
    libpango-1.0-0 \
    libcairo2 \
    libasound2"; fi) \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Copy Playwright dependencies only if not using Browserless
COPY --from=playwright-deps /opt/browsers /opt/browsers
COPY --from=playwright-deps /root/.cache/ms-playwright-go /opt/ms-playwright-go

RUN if [ "$USE_BROWSERLESS" != "true" ]; then \
        chmod -R 755 /opt/browsers && \
        chmod -R 755 /opt/ms-playwright-go; \
    fi

COPY --from=builder /usr/bin/google-maps-scraper /usr/bin/

ENTRYPOINT ["google-maps-scraper"]
