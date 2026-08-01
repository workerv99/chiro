# Chiro

Control de gastos personal. SPA offline-first (SvelteKit + adapter-static) con backend en Go (chi + pgx) y PostgreSQL. Datos personales solo: cada usuario registra sus gastos, categorías, cuentas, presupuestos, préstamos, alcancías y facturas.

## Estructura

| Carpeta | Descripción |
|---------|-------------|
| `web/` | SPA Svelte 5 (SvelteKit + Vite + adapter-static). Salida en `web/build`. |
| `backend/` | API Go (chi, JWT, pgx). Sirve también `web/build` como SPA con fallback. |
| `docker-compose.yml` | PostgreSQL local (`chiro-db`, puerto 5432). |
| `backend/Dockerfile` | Build multi-stage para producción (web + API en una imagen). |

## Requisitos

- Node.js 20+ y npm
- Go 1.23+
- Docker (para la base de datos local)

## Puesta en marcha (desarrollo)

```bash
# 1. Base de datos
docker compose up -d

# 2. Backend (la API espera la DB en postgres://chiro:chiro@localhost:5432/chiro)
cd backend
go run ./cmd/server

# 3. Web (dev server con HMR en http://localhost:5173)
cd web
npm install
npm run dev
```

Variables de entorno: copiar `.env.example` a `.env` (raíz) o `backend/.env.example` a `backend/.env`. En producción el `Dockerfile` multi-stage compila todo y sirve el SPA embebido junto a la API en el puerto 8080.

## Comandos web

```bash
npm run dev        # dev server
npm run build      # build de producción (web/build)
npm run preview    # previsualizar el build
npm run check      # svelte-check (tipos)
npm run lint       # ESLint (config plano, reglas svelte recomendadas)
```

## API

La API expone rutas `/api/*` (auth JWT Bearer). Las rutas de datos son por-usuario:

- `POST /api/auth/register`, `POST /api/auth/login`, `GET /api/auth/me`
- `POST /api/sync` — sincronización incremental por tabla (`since`/`tombstones`)
- `GET /api/expenses`, `POST /api/expenses`, `PUT/DELETE /api/expenses/{id}`
- Categorías, cuentas, etiquetas, presupuestos, préstamos (con cuotas y cascada), alcancías y facturas.

## Migraciones

Las migraciones SQL van embebidas en el binario. Para aplicarlas:

```bash
cd backend
DATABASE_URL=postgres://chiro:chiro@localhost:5432/chiro go run ./cmd/migrate
```
