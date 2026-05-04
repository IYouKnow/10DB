# 10DB Launch

Self-hosted visual PostgreSQL launchpad for creating isolated project databases, designing schemas, previewing SQL, and copying ready-to-use connection strings.

## Stack

- Go + Chi + pgx + SQLite (`modernc.org/sqlite`)
- React + Vite + TypeScript
- Tailwind CSS
- React Flow
- Docker Compose

## Docker / CasaOS

The repo now includes a root `docker-compose.yml` and `Dockerfile` so you can run it directly in Docker or import it into CasaOS.

1. Copy `.env.example` to `.env`
2. Set at least:
   - `APP_MASTER_KEY`
   - `PG_ADMIN_PASSWORD`
   - `APP_BASE_URL`
   - `APP_ALLOWED_ORIGINS`
3. Start the stack:

```bash
docker compose up -d --build
```

The app will be available on `http://localhost:8080`.

Notes:
- The app stores its control SQLite DB in the `app_data` volume at `/data/10db-launch.sqlite`
- PostgreSQL data is stored in the `postgres_data` volume
- Inside Docker, the app automatically connects to PostgreSQL using `PG_ADMIN_HOST=postgres`

## Git Tags

To publish a tagged release to GitHub from `main`:

```bash
git add .
git commit -m "Add Docker Compose deployment"
git tag v0.1.0
git push origin main
git push origin v0.1.0
```

If you want to push all local tags instead:

```bash
git push origin --tags
```
