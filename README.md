# Chiro

Control de gastos personal. SPA offline-first (SvelteKit + adapter-static) con backend en Go (chi + pgx) y PostgreSQL. Datos personales solo: cada usuario registra sus gastos, categorías, cuentas, presupuestos, préstamos, alcancías y facturas.

## Estructura

| Carpeta | Descripción |
|---------|-------------|
| `web/` | SPA Svelte 5 (SvelteKit + Vite + adapter-static). Salida en `web/build`. |
| `backend/` | API Go (chi, JWT, pgx). Sirve también `web/build` como SPA con fallback. |
| `prisma/` | Tooling de esquema: `schema.prisma` espeja el SQL de migraciones; se usa solo para `db push`/`studio` (el runtime NO usa Prisma Client). |
| `docker-compose.yml` | PostgreSQL local (`chiro-db`, puerto 5432) para desarrollo sin red. |
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

## Base de datos

Dos modos, intercambiando `DATABASE_URL`:

- **Local**: `docker compose up -d` → `postgres://chiro:chiro@localhost:5432/chiro?sslmode=disable`.
- **Vercel Postgres (Prisma Postgres)**: las connection strings van en el `.env` de la raíz (`DATABASE_URL`, `POSTGRES_URL`, `PRISMA_DATABASE_URL`). El backend Go (`pgx`) usa `DATABASE_URL` tal cual; solo cambia el valor. El host de Vercel exige `sslmode=require` (ya viene en la URL).

### Esquema con Prisma

`prisma/schema.prisma` es la fuente del esquema en la nube (espejo de `backend/internal/migrate/migrations/001_init.sql`: tablas con `updated_at` BIGINT ms + `deleted` para sincronización LWW, PKs compuestas `(user_id, id)`). El runtime sigue siendo Go — Prisma se usa solo como herramienta:

```bash
cd prisma
npm install
npx prisma validate       # valida el schema
npx prisma db push        # aplica el esquema a la DB (Vercel o local vía PRISMA_DATABASE_URL)
npx prisma studio         # explorador web de datos
```

`prisma/.env` guarda solo `PRISMA_DATABASE_URL`. El backend Go NO usa Prisma Client; ejecuta sus propias migraciones idempotentes (`CREATE TABLE IF NOT EXISTS`) al arrancar.

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
- `GET/PUT /api/admin/users` y `GET /api/admin/users/{id}` — solo rol `admin`.

## Cuentas y roles (SaaS)

Cada usuario se crea solo (`role='user'`, `status='active'`). El admin es un usuario con `role='admin'` que gestiona cuentas:

- **JWT** lleva `role`; el middleware `requireAdmin` protege `/api/admin/*`, y `requireActive` rechaza cuentas `disabled` en cada petición (aunque el token siga vigente).
- **Admin puede**: listar usuarios, ver detalle con conteos, activar/desactivar, resetear contraseña y cambiar rol. No puede quitarse el rol admin ni desactivarse a sí mismo (anti-lockout).
- **Primer admin**:

  ```bash
  ADMIN_EMAIL=admin@dominio ADMIN_PASSWORD=secreta go run ./cmd/admin-create
  ```

  Si el email ya existe, lo promueve a admin.

## Importar datos de un usuario (SQLite → Postgres)

Cada `gastos.db` del app original se importa como la cuenta de un usuario. El comando copia las 12 tablas respetando `updated_at`/`deleted` (merge LWW) y valida las fechas:

```bash
# Bajo un usuario existente:
go run ./cmd/import-sqlite -db gastos_backup.db -user-id usr_xxx

# O creando el usuario a la vez:
go run ./cmd/import-sqlite -db gastos_backup.db -email user@dominio -password secreta -name "Nombre"
```

Si una fila trae una fecha inválida (p.ej. `0263-42-00`), aborta indicando la fila para corregirla en el `.db` y volver a ejecutar (es idempotente).

## Migraciones

Las migraciones SQL van embebidas en el binario. Para aplicarlas:

```bash
cd backend
DATABASE_URL=postgres://chiro:chiro@localhost:5432/chiro go run ./cmd/migrate
```
