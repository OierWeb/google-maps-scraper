# Integración con Browserless

Este documento explica cómo configurar el Google Maps Scraper para usar Browserless como servicio remoto de navegador.

## ¿Qué es Browserless?

Browserless es un servicio que proporciona navegadores remotos a través de WebSocket, permitiendo ejecutar Playwright y Puppeteer sin necesidad de instalar navegadores localmente.

## Configuración

### 1. Variable de Entorno

Para usar Browserless, simplemente configura la variable de entorno `BROWSER_WS_ENDPOINT`:

```bash
export BROWSER_WS_ENDPOINT="ws://browserless-voc40o8w8ccko?token=3BBYUlCr1FLPDIrMgE"
```

O en Windows:
```cmd
set BROWSER_WS_ENDPOINT=ws://browserless-voc40o8w8ccko?token=3BBYUlCr1FLPDIrMgE
```

### 2. Formatos de Endpoint Soportados

El scraper soporta varios formatos de endpoint:

- `ws://localhost:3000` (desarrollo local)
- `ws://browserless-instance?token=your-token` (instancia remota)
- `wss://production-sfo.browserless.io?token=your-token` (producción)

**Nota**: El scraper automáticamente convierte `ws://` a `wss://` para conexiones remotas (no localhost).

## Uso

### Modo Archivo (File Runner)

```bash
# Con Browserless
export BROWSER_WS_ENDPOINT="ws://your-browserless-instance?token=your-token"
./google-maps-scraper -input queries.txt -results results.csv

# Sin Browserless (navegador local)
unset BROWSER_WS_ENDPOINT
./google-maps-scraper -input queries.txt -results results.csv
```

### Modo Web

```bash
export BROWSER_WS_ENDPOINT="ws://your-browserless-instance?token=your-token"
./google-maps-scraper -web -addr :8080
```

### Modo Base de Datos

```bash
export BROWSER_WS_ENDPOINT="ws://your-browserless-instance?token=your-token"
./google-maps-scraper -dsn "postgres://user:pass@localhost/db"
```

## Ventajas de Usar Browserless

1. **Sin instalación de navegadores**: No necesitas instalar Chrome/Chromium localmente
2. **Escalabilidad**: Puedes usar múltiples instancias de Browserless
3. **Recursos optimizados**: El navegador se ejecuta en un servidor dedicado
4. **Mejor estabilidad**: Browserless está optimizado para scraping
5. **Compatibilidad con Docker**: Perfecto para contenedores sin GUI

## Configuración Automática

Cuando se detecta `BROWSER_WS_ENDPOINT`, el scraper automáticamente:

- ✅ Configura la conexión remota a Browserless
- ✅ Desactiva la descarga de navegadores locales
- ✅ Optimiza timeouts para conexiones remotas
- ✅ Aplica configuraciones específicas para Browserless
- ✅ Usa el scroll mejorado con mejor manejo de errores

## Ejemplo de Docker Compose

```yaml
version: '3.8'
services:
  browserless:
    image: browserless/chrome:latest
    ports:
      - "3000:3000"
    environment:
      - TOKEN=your-secure-token
      - MAX_CONCURRENT_SESSIONS=10
      - CONNECTION_TIMEOUT=60000

  google-maps-scraper:
    build: .
    environment:
      - BROWSER_WS_ENDPOINT=ws://browserless:3000?token=your-secure-token
    volumes:
      - ./queries.txt:/app/queries.txt
      - ./results:/app/results
    command: ["-input", "queries.txt", "-results", "results/output.csv"]
    depends_on:
      - browserless
```

## Troubleshooting

### Error de Conexión

Si ves errores como "connection refused":

1. Verifica que Browserless esté ejecutándose
2. Confirma que el token sea correcto
3. Revisa que el puerto esté accesible

### Timeouts

El scraper usa timeouts aumentados para conexiones remotas:

- Navegación: 30 segundos (vs 5 segundos local)
- Elementos: 15 segundos (vs 5 segundos local)
- Scroll: 90-180 segundos por iteración

### Logs

Cuando uses Browserless, verás logs como:

```
🌐 Configuring Browserless connection to: ws://your-endpoint
🚀 Using Browserless remote browser configuration
```

## Configuración Avanzada

### Variables de Entorno Adicionales

```bash
# Endpoint de Browserless
export BROWSER_WS_ENDPOINT="ws://browserless:3000?token=your-token"

# Desactivar descarga de navegadores (automático con Browserless)
export PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1

# Configuración de Playwright para Browserless
export PLAYWRIGHT_BROWSERS_PATH=0
```

### Parámetros de Línea de Comandos

Todos los parámetros normales funcionan con Browserless:

```bash
./google-maps-scraper \
  -input queries.txt \
  -results results.csv \
  -c 5 \
  -depth 20 \
  -lang es \
  -email \
  -json
```

## Compatibilidad

- ✅ File Runner
- ✅ Web Runner  
- ✅ Database Runner
- ✅ Lambda AWS Runner
- ✅ Todos los modos de scraping
- ✅ Extracción de emails
- ✅ Coordenadas geográficas
- ✅ Múltiples idiomas

## Rendimiento

Con Browserless puedes esperar:

- **Latencia**: +100-300ms por conexión remota
- **Throughput**: Similar al navegador local
- **Memoria**: Reducida en el cliente
- **CPU**: Reducida en el cliente
- **Estabilidad**: Mejorada con retry logic

## Seguridad

- Usa tokens seguros para Browserless
- Considera usar `wss://` para conexiones en producción
- No hardcodees tokens en el código
- Usa variables de entorno o archivos de configuración seguros
