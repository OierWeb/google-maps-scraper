# Requirements Document

## Introduction

Este proyecto actualmente descarga e instala Playwright con Chromium localmente para realizar scraping de Google Maps. El objetivo es migrar la configuración para usar una instancia existente de Browserless que ya está desplegada en la misma red Docker, eliminando la necesidad de descargar Chromium localmente y aprovechando los recursos compartidos.

## Requirements

### Requirement 1

**User Story:** Como desarrollador del sistema, quiero conectar el scraper a la instancia existente de Browserless, para que no necesite descargar Chromium localmente y pueda aprovechar los recursos compartidos.

#### Acceptance Criteria

1. WHEN el sistema inicia THEN debe conectarse a la instancia de Browserless en lugar de usar Playwright local
2. WHEN se ejecuta un job de scraping THEN debe usar el navegador remoto de Browserless
3. WHEN se configura la conexión THEN debe usar las credenciales y URL de la instancia de Browserless existente

### Requirement 2

**User Story:** Como administrador del sistema, quiero configurar la conexión a Browserless mediante variables de entorno, para que pueda cambiar la configuración sin modificar el código.

#### Acceptance Criteria

1. WHEN se proporciona BROWSERLESS_URL THEN el sistema debe usar esa URL para conectarse
2. WHEN se proporciona BROWSERLESS_TOKEN THEN el sistema debe usar ese token para autenticación
3. IF no se proporcionan las variables de entorno THEN el sistema debe usar valores por defecto o fallar graciosamente

### Requirement 3

**User Story:** Como usuario del sistema, quiero que el comportamiento del scraping sea idéntico al actual, para que no haya cambios en la funcionalidad existente.

#### Acceptance Criteria

1. WHEN se ejecuta un scraping job THEN debe producir los mismos resultados que antes
2. WHEN se usan proxies THEN deben funcionar correctamente con Browserless
3. WHEN se configuran opciones de navegador THEN deben aplicarse correctamente en Browserless

### Requirement 4

**User Story:** Como desarrollador, quiero eliminar la dependencia de instalación local de Playwright/ chromiun que ya no sean necesarias, sin romper nada de funcionalidad, para que el contenedor Docker sea más ligero y el despliegue más rápido.

#### Acceptance Criteria

1. WHEN se construye el contenedor Docker THEN no debe descargar Chromium
2. WHEN se ejecuta el modo de instalación de Playwright THEN debe ser obsoleto o redirigido
3. WHEN se inicia la aplicación THEN no debe requerir binarios de Playwright locales