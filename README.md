# 10DB Launch

Self-hosted visual PostgreSQL launchpad for creating isolated project databases, designing schemas, previewing SQL, and copying ready-to-use connection strings.

## Stack

- Go + Chi + pgx + SQLite (`modernc.org/sqlite`)
- React + Vite + TypeScript
- Tailwind CSS
- React Flow
- Docker Compose

## Docker / CasaOS

The repo now includes:
- a root `Dockerfile` for building the app image
- a root `docker-compose.yml` for running only the `10db` app container
- a GitHub Actions workflow that publishes the image automatically

`10db` expects an existing PostgreSQL server. It does not need to run PostgreSQL in the same Compose file.

1. Copy `.env.example` to `.env`
2. Set at least:
   - `APP_MASTER_KEY`
   - `PG_ADMIN_HOST`
   - `PG_ADMIN_PASSWORD`
   - `APP_BASE_URL`
   - `APP_ALLOWED_ORIGINS`
3. Start the stack:

```bash
docker compose up -d
```

The app will be available on `http://localhost:8080`.

Notes:
- The app stores its control SQLite DB in the `app_data` volume at `/data/10db-launch.sqlite`
- PostgreSQL is external to this Compose file
- In CasaOS, import or configure only the `10db` app container and point `PG_ADMIN_HOST` to your existing PostgreSQL server
- By default, Compose pulls `ghcr.io/iyouknow/10db:v0.1.0`
- To use another published tag, set `APP_IMAGE`, for example `APP_IMAGE=ghcr.io/iyouknow/10db:latest`

## GitHub Actions Image Publishing

The workflow at `.github/workflows/docker-publish.yml` publishes the app image to `ghcr.io/iyouknow/10db`.

It runs on:
- pushes to `main`
- pushes of tags like `v0.1.0`
- manual runs from the Actions tab

Published tags include:
- `latest` for the default branch
- the git tag name, like `v0.1.0`
- a `sha-...` tag for traceability

For CasaOS, use the published image instead of a source build.

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
