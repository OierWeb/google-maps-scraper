# Browserless Deployment Guide

This guide explains how to deploy Google Maps Scraper using Browserless for remote browser functionality, eliminating the need for local Chromium installation.

## Overview

Browserless provides a remote browser service that the scraper can connect to via WebSocket. This approach offers several advantages:

- **Reduced Resource Usage**: No need to download and install Chromium locally
- **Better Scaling**: Share browser instances across multiple scraper processes  
- **Faster Deployment**: Smaller Docker images without browser binaries
- **Centralized Management**: Single Browserless instance can serve multiple scrapers

## Quick Start

### Option 1: Using docker-compose (Recommended)

The easiest way to get started is using the provided docker-compose configuration:

```bash
# Clone the repository
git clone https://github.com/gosom/google-maps-scraper.git
cd google-maps-scraper

# Start all services (Browserless + Scraper + Database)
docker-compose -f docker-compose.browserless.yaml up -d

# Check service status
docker-compose -f docker-compose.browserless.yaml ps
```

### Option 2: Manual Docker Setup

1. **Start Browserless service:**
```bash
docker run -d \
  --name browserless \
  -p 3000:3000 \
  -e TOKEN=your-secure-token \
  -e CONCURRENT=10 \
  -e MAX_CONCURRENT_SESSIONS=10 \
  browserless/chrome:latest
```

2. **Build scraper with Browserless support:**
```bash
docker build --build-arg USE_BROWSERLESS=true -t gmaps-scraper:browserless .
```

3. **Run scraper with Browserless:**
```bash
docker run --link browserless \
  -e USE_BROWSERLESS=true \
  -e BROWSERLESS_URL=ws://browserless:3000 \
  -e BROWSERLESS_TOKEN=your-secure-token \
  -v $PWD/example-queries.txt:/example-queries \
  -v $PWD/results.csv:/results.csv \
  gmaps-scraper:browserless \
  -depth 1 -input /example-queries -results /results.csv -exit-on-inactivity 3m
```

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `USE_BROWSERLESS` | Enable Browserless mode | `false` | Yes |
| `BROWSERLESS_URL` | WebSocket URL to Browserless | `ws://browserless:3000` | Yes |
| `BROWSERLESS_TOKEN` | Authentication token | - | Yes |

### Browserless Configuration

Configure Browserless service with these environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `TOKEN` | Authentication token | - |
| `CONCURRENT` | Max concurrent browser instances | `10` |
| `MAX_CONCURRENT_SESSIONS` | Max concurrent sessions | `10` |
| `QUEUE_LENGTH` | Queue length for requests | `50` |
| `MAX_MEMORY_PERCENT` | Max memory usage percentage | `95` |
| `MAX_CPU_PERCENT` | Max CPU usage percentage | `95` |
| `DEBUG` | Enable debug mode | `false` |

## Docker Compose Examples

### Basic Setup (docker-compose.browserless.yaml)

```yaml
services:
  browserless:
    image: browserless/chrome:latest
    ports:
      - '3000:3000'
    environment:
      - TOKEN=your-secure-token
      - CONCURRENT=10
      - MAX_CONCURRENT_SESSIONS=10
    restart: unless-stopped

  gmaps-scraper:
    build: 
      context: .
      args:
        USE_BROWSERLESS: "true"
    environment:
      - USE_BROWSERLESS=true
      - BROWSERLESS_URL=ws://browserless:3000
      - BROWSERLESS_TOKEN=your-secure-token
    depends_on:
      - browserless
    restart: unless-stopped
```

### Production Setup with Database

```yaml
services:
  browserless:
    image: browserless/chrome:latest
    environment:
      - TOKEN=${BROWSERLESS_TOKEN}
      - CONCURRENT=20
      - MAX_CONCURRENT_SESSIONS=20
    deploy:
      resources:
        limits:
          memory: 4G
          cpus: '2.0'
    restart: unless-stopped

  gmaps-scraper:
    build: 
      context: .
      args:
        USE_BROWSERLESS: "true"
    environment:
      - USE_BROWSERLESS=true
      - BROWSERLESS_URL=ws://browserless:3000
      - BROWSERLESS_TOKEN=${BROWSERLESS_TOKEN}
      - DATABASE_URL=${DATABASE_URL}
    depends_on:
      - browserless
      - db
    deploy:
      replicas: 3
    restart: unless-stopped

  db:
    image: postgres:15.2-alpine
    environment:
      - POSTGRES_USER=${POSTGRES_USER}
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped
```

## Kubernetes Deployment

### Browserless Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: browserless
spec:
  replicas: 2
  selector:
    matchLabels:
      app: browserless
  template:
    metadata:
      labels:
        app: browserless
    spec:
      containers:
      - name: browserless
        image: browserless/chrome:latest
        ports:
        - containerPort: 3000
        env:
        - name: TOKEN
          valueFrom:
            secretKeyRef:
              name: browserless-secret
              key: token
        - name: CONCURRENT
          value: "15"
        - name: MAX_CONCURRENT_SESSIONS
          value: "15"
        resources:
          limits:
            memory: "2Gi"
            cpu: "1000m"
          requests:
            memory: "1Gi"
            cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: browserless-service
spec:
  selector:
    app: browserless
  ports:
  - port: 3000
    targetPort: 3000
```

### Scraper Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gmaps-scraper
spec:
  replicas: 5
  selector:
    matchLabels:
      app: gmaps-scraper
  template:
    metadata:
      labels:
        app: gmaps-scraper
    spec:
      containers:
      - name: gmaps-scraper
        image: gmaps-scraper:browserless
        env:
        - name: USE_BROWSERLESS
          value: "true"
        - name: BROWSERLESS_URL
          value: "ws://browserless-service:3000"
        - name: BROWSERLESS_TOKEN
          valueFrom:
            secretKeyRef:
              name: browserless-secret
              key: token
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: url
        args: ["-c", "2", "-depth", "10", "-dsn", "$(DATABASE_URL)"]
        resources:
          limits:
            memory: "512Mi"
            cpu: "500m"
          requests:
            memory: "256Mi"
            cpu: "250m"
```

## Performance Tuning

### Browserless Configuration

For high-throughput scenarios, adjust these Browserless settings:

```bash
# High performance setup
docker run -d \
  --name browserless \
  -p 3000:3000 \
  -e TOKEN=your-secure-token \
  -e CONCURRENT=25 \
  -e MAX_CONCURRENT_SESSIONS=25 \
  -e QUEUE_LENGTH=100 \
  -e MAX_MEMORY_PERCENT=90 \
  -e MAX_CPU_PERCENT=90 \
  --memory=4g \
  --cpus=2 \
  browserless/chrome:latest
```

### Scraper Configuration

Run multiple scraper instances to maximize throughput:

```bash
# Scale scraper instances
docker-compose -f docker-compose.browserless.yaml up -d --scale gmaps-scraper=5
```

## Monitoring and Troubleshooting

### Health Checks

Browserless provides a health endpoint:

```bash
curl http://localhost:3000/health
```

### Logs

Monitor Browserless logs:

```bash
docker logs -f browserless
```

Monitor scraper logs:

```bash
docker-compose -f docker-compose.browserless.yaml logs -f gmaps-scraper
```

### Common Issues

1. **Connection Refused**: Ensure Browserless is running and accessible
2. **Authentication Failed**: Verify BROWSERLESS_TOKEN matches
3. **Resource Limits**: Increase Browserless memory/CPU limits if needed
4. **Network Issues**: Check Docker network connectivity between services

### Debug Mode

Enable debug mode for troubleshooting:

```bash
# Browserless debug
docker run -e DEBUG=true browserless/chrome:latest

# Scraper debug
docker run -e LOG_LEVEL=debug gmaps-scraper:browserless
```

## Security Considerations

1. **Token Security**: Use strong, unique tokens for Browserless authentication
2. **Network Security**: Run Browserless on internal networks only
3. **Resource Limits**: Set appropriate memory and CPU limits
4. **Access Control**: Restrict access to Browserless endpoints

## Migration from Local Playwright

To migrate existing deployments:

1. **Update Environment Variables**: Add Browserless configuration
2. **Rebuild Images**: Use `USE_BROWSERLESS=true` build arg
3. **Update Compose Files**: Add Browserless service
4. **Test Functionality**: Verify scraping works with remote browser
5. **Monitor Performance**: Compare performance metrics

## Cost Optimization

- **Shared Browserless**: Use one Browserless instance for multiple scrapers
- **Resource Scaling**: Scale Browserless based on concurrent scraper instances
- **Efficient Queuing**: Configure appropriate queue lengths to avoid resource waste
- **Auto-scaling**: Use Kubernetes HPA for dynamic scaling based on load

## Support

For issues related to:
- **Browserless**: Check [Browserless documentation](https://docs.browserless.io/)
- **Google Maps Scraper**: Open an issue on the [GitHub repository](https://github.com/gosom/google-maps-scraper)