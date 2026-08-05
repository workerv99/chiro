# Chiro

SaaS de control de gastos personales. SPA (SvelteKit + adapter-static) con backend en Go (chi + pgx) y PostgreSQL. Los usuarios registran gastos, categorías, cuentas, presupuestos, préstamos y facturas.

## Características

- **Gastos**: Registro de ingresos/egresos con categorías y etiquetas
- **Presupuestos**: Límites por categoría con progreso visual
- **Préstamos**: Calendario de cuotas con pagos parciales y cascada
- **Reportes PDF**: Generación de reportes profesionales de préstamos
- **Multi-idioma**: Español e inglés (i18n)
- **Dark mode**: Tema AMOLED por defecto (alternativa clara)
- **Landing page**: Página de marketing con planes Free/Pro
- **Onboarding**: Wizard guiado para nuevos usuarios
- **GDPR**: Consentimiento, exportación y eliminación de datos
- **Admin**: Dashboard con métricas y gestión de usuarios

## Planes

| Plan | Precio | Límites |
|------|--------|---------|
| Free | $0/mes | 50 gastos, 3 cuentas, 10 préstamos |
| Pro | $4.99/mes | Ilimitado |

## Estructura

| Carpeta | Descripción |
|---------|-------------|
| `web/` | SPA Svelte 5 (SvelteKit + Vite + adapter-static) |
| `backend/` | API Go (chi, JWT, pgx) |
| `prisma/` | Tooling de esquema (no usado en runtime) |
| `docker-compose.yml` | PostgreSQL local para desarrollo |

## Requisitos

- Node.js 20+ y npm
- Go 1.23+
- Docker (para la base de datos local)

## Desarrollo

```bash
# 1. Base de datos
docker compose up -d

# 2. Backend
cd backend
go run ./cmd/server

# 3. Web (dev server en http://localhost:5173)
cd web
npm install
npm run dev
```

## Variables de entorno

```bash
# Backend
DATABASE_URL=postgres://...
JWT_SECRET=tu-secreto
RESEND_API_KEY=re_...        # Email transaccional (opcional)
FROM_EMAIL=noreply@chiro.app

# Frontend
VITE_API_BASE_URL=https://chiro-backend.vercel.app
```

## Despliegue

### Backend (Vercel)
1. Crear proyecto en Vercel apuntando a `backend/`
2. Agregar `DATABASE_URL` y `JWT_SECRET` en Variables de Entorno
3. El `api/index.go` es el entry point para Vercel

### Frontend (Cloudflare Pages)
1. Crear proyecto en Cloudflare Pages apuntando a `web/`
2. Agregar `VITE_API_BASE_URL` apuntando al backend
3. Build command: `npm run build`
4. Output directory: `build`

### Base de datos (Supabase)
1. Crear proyecto en Supabase
2. Usar **Transaction Pooler** (IPv4 compatible con Vercel)
3. Agregar `?sslmode=require&pgbouncer=true` a la URL
4. Ejecutar migraciones: `go run ./cmd/migrate`

## API

### Auth
- `POST /api/auth/register` — Registro (requiere terms_accepted + privacy_accepted)
- `POST /api/auth/login` — Login
- `GET /api/auth/me` — Usuario actual

### Datos (por usuario)
- `GET/POST /api/expenses` — Gastos
- `GET/POST /api/accounts` — Cuentas
- `GET/POST /api/categories` — Categorías
- `GET/POST /api/loans` — Préstamos
- `GET/POST /api/budgets` — Presupuestos
- `GET/POST /api/bills` — Facturas recurrentes
- `POST /api/sync` — Sincronización incremental

### Préstamos
- `POST /api/loans` — Crear préstamo
- `PUT /api/loans/{id}` — Actualizar y regenerar cronograma
- `POST /api/installments/{id}/pay` — Pagar cuota
- `POST /api/installments/{id}/cascade` — Pago con cascada (excedente a siguientes cuotas)
- `POST /api/installments/{id}/unpay` — Deshacer último pago

### Suscripciones
- `GET /api/subscription` — Estado del plan
- `POST /api/subscription/activate` — Activar Pro
- `POST /api/subscription/cancel` — Cancelar suscripción

### GDPR
- `DELETE /api/account` — Eliminar cuenta y todos los datos
- `GET /api/export` — Exportar todos los datos en JSON

### Admin (solo rol admin)
- `GET /api/admin/stats` — Métricas generales
- `GET /api/admin/users` — Listar usuarios
- `GET /api/admin/users/{id}` — Detalle de usuario
- `PUT /api/admin/users/{id}` — Actualizar usuario

## Migraciones

```bash
cd backend
DATABASE_URL=postgres://... go run ./cmd/migrate
```

Las migraciones se ejecutan automáticamente al iniciar el servidor.

## Importar datos (SQLite → Postgres)

```bash
# Bajo un usuario existente:
go run ./cmd/import-sqlite -db gastos_backup.db -user-id usr_xxx

# Creando el usuario:
go run ./cmd/import-sqlite -db gastos_backup.db -email user@dominio -password secreta -name "Nombre"
```

## Licencia

MIT
