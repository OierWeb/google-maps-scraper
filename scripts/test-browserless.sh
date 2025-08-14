#!/bin/bash

# Script para probar la integración con Browserless
# Uso: ./scripts/test-browserless.sh

set -e

echo "🧪 Testing Google Maps Scraper with Browserless Integration"
echo "=========================================================="

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Función para imprimir mensajes coloreados
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Verificar si la variable BROWSER_WS_ENDPOINT está configurada
if [ -z "$BROWSER_WS_ENDPOINT" ]; then
    print_error "BROWSER_WS_ENDPOINT no está configurada"
    echo ""
    echo "Por favor, configura la variable de entorno:"
    echo "export BROWSER_WS_ENDPOINT=\"ws://browserless-voc40o8w8ccko?token=3BBYUlCr1FLPDIrMgE\""
    echo ""
    echo "O para desarrollo local:"
    echo "export BROWSER_WS_ENDPOINT=\"ws://localhost:3000\""
    exit 1
fi

print_success "BROWSER_WS_ENDPOINT configurada: $BROWSER_WS_ENDPOINT"

# Crear archivo de consultas de prueba
TEST_QUERIES="test_queries.txt"
print_status "Creando archivo de consultas de prueba: $TEST_QUERIES"

cat > $TEST_QUERIES << EOF
restaurantes madrid
hoteles barcelona
farmacias valencia
EOF

print_success "Archivo de consultas creado con 3 consultas de prueba"

# Compilar el proyecto
print_status "Compilando el proyecto..."
if go build -o google-maps-scraper .; then
    print_success "Compilación exitosa"
else
    print_error "Error en la compilación"
    exit 1
fi

# Ejecutar prueba con Browserless
print_status "Ejecutando scraper con Browserless..."
print_status "Configuración:"
echo "  - Endpoint: $BROWSER_WS_ENDPOINT"
echo "  - Consultas: $TEST_QUERIES"
echo "  - Resultados: browserless_results.csv"
echo "  - Concurrencia: 2"
echo "  - Profundidad: 5"

# Ejecutar el scraper
if ./google-maps-scraper \
    -input "$TEST_QUERIES" \
    -results "browserless_results.csv" \
    -c 2 \
    -depth 5 \
    -lang es; then
    
    print_success "Scraping completado exitosamente"
    
    # Verificar resultados
    if [ -f "browserless_results.csv" ]; then
        RESULT_COUNT=$(wc -l < "browserless_results.csv")
        print_success "Archivo de resultados creado con $RESULT_COUNT líneas"
        
        echo ""
        print_status "Primeras 5 líneas de resultados:"
        head -5 "browserless_results.csv"
    else
        print_warning "No se encontró el archivo de resultados"
    fi
    
else
    print_error "Error durante el scraping"
    exit 1
fi

# Limpiar archivos de prueba
print_status "Limpiando archivos de prueba..."
rm -f "$TEST_QUERIES"

print_success "✅ Prueba de integración con Browserless completada exitosamente"

echo ""
echo "📊 Resumen:"
echo "  - Browserless endpoint: $BROWSER_WS_ENDPOINT"
echo "  - Consultas procesadas: 3"
echo "  - Archivo de resultados: browserless_results.csv"
echo ""
echo "🎉 La integración con Browserless está funcionando correctamente!"
