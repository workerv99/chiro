#!/bin/bash
set -e

echo "🚀 Iniciando Chiro en modo desarrollo..."

# Iniciar Docker (DB + Backend)
echo "📦 Iniciando Docker (PostgreSQL + Backend)..."
docker compose up -d db backend

# Esperar a que PostgreSQL esté listo
echo "⏳ Esperando a que PostgreSQL esté listo..."
sleep 5

# Verificar que el backend esté corriendo
echo "🔍 Verificando backend..."
docker compose logs backend --tail=10

echo ""
echo "✅ Servicios iniciados:"
echo "   - PostgreSQL: localhost:5432"
echo "   - Backend API: http://localhost:8080"
echo ""
echo "💡 Para iniciar el frontend (fuera de Docker):"
echo "   cd web && npm run dev"
echo ""
echo "📋 Comandos útiles:"
echo "   docker compose logs -f backend    # Ver logs del backend"
echo "   docker compose down               # Detener todo"
echo "   docker compose restart backend    # Reiniciar backend"
