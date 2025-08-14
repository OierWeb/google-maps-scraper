# Design Document

## Overview

La migración de Playwright local a Browserless requiere modificar la configuración de scrapemate para conectarse a una instancia remota de navegador en lugar de usar una instalación local. Browserless proporciona una API compatible con Playwright que permite conectarse mediante WebSocket.

Según la documentación de Browserless, la conexión se realiza mediante el endpoint WebSocket: `ws://browserless-host:3000` con autenticación por token.

## Architecture

### Current Architecture
```
Google Maps Scraper → scrapemate → Playwright (local) → Chromium (local)
```

### Target Architecture
```
Google Maps Scraper → scrapemate → Playwright (remote) → Browserless → Chromium (remote)
```

### Configuration Flow
1. El sistema lee variables de entorno para configuración de Browserless
2. scrapemate se configura con el endpoint remoto de Browserless
3. Playwright se conecta al navegador remoto mediante WebSocket
4. Se elimina la necesidad de instalación local de Playwright/Chromium

## Components and Interfaces

### 1. Configuration Management
- **Location**: `runner/runner.go`
- **Purpose**: Agregar nuevos campos de configuración para Browserless
- **New Fields**:
  - `BrowserlessURL`: URL del servicio Browserless
  - `BrowserlessToken`: Token de autenticación
  - `UseBrowserless`: Flag para habilitar/deshabilitar Browserless

### 2. Scrapemate Configuration
- **Location**: Todos los runners (`filerunner`, `webrunner`, `databaserunner`, `lambdaaws`)
- **Purpose**: Modificar la configuración de scrapemate para usar endpoint remoto
- **Changes**: Agregar configuración de navegador remoto cuando Browserless esté habilitado

### 3. Playwright Installation Handler
- **Location**: `runner/installplaywright/installplaywright.go`
- **Purpose**: Modificar o deshabilitar la instalación cuando se usa Browserless
- **Changes**: Saltar instalación si Browserless está configurado

### 4. Docker Configuration
- **Location**: `Dockerfile`
- **Purpose**: Eliminar la instalación de Chromium del contenedor
- **Changes**: Remover pasos de instalación de Playwright/Chromium

## Data Models

### Configuration Extension
```go
type Config struct {
    // ... existing fields ...
    BrowserlessURL   string
    BrowserlessToken string
    UseBrowserless   bool
}
```

### Environment Variables
- `BROWSERLESS_URL`: URL del servicio Browserless (default: "ws://browserless:3000")
- `BROWSERLESS_TOKEN`: Token de autenticación
- `USE_BROWSERLESS`: Flag para habilitar Browserless (default: "false")

## Error Handling

### Connection Errors
- **Scenario**: Browserless no disponible
- **Handling**: Fallback a Playwright local si está disponible, o error claro
- **Logging**: Log detallado de intentos de conexión

### Authentication Errors
- **Scenario**: Token inválido o expirado
- **Handling**: Error inmediato con mensaje claro
- **Logging**: Log de errores de autenticación (sin exponer token)

### Network Errors
- **Scenario**: Problemas de red con Browserless
- **Handling**: Reintentos con backoff exponencial
- **Logging**: Log de problemas de conectividad

## Testing Strategy

### Unit Tests
- Configuración de variables de entorno
- Validación de parámetros de conexión
- Manejo de errores de configuración

### Integration Tests
- Conexión exitosa a Browserless
- Ejecución de jobs de scraping con navegador remoto
- Comparación de resultados entre local y remoto

### End-to-End Tests
- Scraping completo de Google Maps usando Browserless
- Verificación de funcionalidad de proxies
- Pruebas de rendimiento comparativo

## Implementation Approach

### Phase 1: Configuration Setup
1. Agregar nuevos campos de configuración
2. Implementar lectura de variables de entorno
3. Agregar validación de configuración

### Phase 2: Scrapemate Integration
1. Investigar API de scrapemate para navegador remoto
2. Implementar configuración de endpoint remoto
3. Modificar todos los runners para usar nueva configuración

### Phase 3: Playwright Installation Bypass
1. Modificar installplaywright para detectar Browserless
2. Saltar instalación cuando Browserless esté habilitado
3. Mantener compatibilidad hacia atrás

### Phase 4: Docker Optimization
1. Remover instalación de Chromium del Dockerfile
2. Optimizar imagen para uso con Browserless
3. Actualizar documentación de despliegue

## Security Considerations

- Token de Browserless debe manejarse como secreto
- Conexión WebSocket debe validar certificados si usa HTTPS
- Logs no deben exponer tokens de autenticación
- Configuración debe permitir conexiones seguras (WSS)

## Performance Considerations

- Conexión remota puede tener latencia adicional
- Pool de conexiones para optimizar rendimiento
- Monitoreo de tiempo de respuesta de Browserless
- Configuración de timeouts apropiados para conexiones remotas