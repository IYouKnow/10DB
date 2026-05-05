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
- a root `docker-compose.yml` formatted for CasaOS import
- a GitHub Actions workflow that publishes the image automatically

`10db` expects an existing PostgreSQL server. It does not need to run PostgreSQL in the same Compose file.

Notes:
- The app stores its control SQLite DB in `/data/10db-launch.sqlite`
- PostgreSQL is external to this Compose file
- The CasaOS compose file uses literal values and `x-casaos` metadata so CasaOS imports it cleanly
- The default image is `ghcr.io/iyouknow/10db:latest`
- Before install, replace the placeholder values for `APP_MASTER_KEY`, `APP_BASE_URL`, `APP_ALLOWED_ORIGINS`, `PG_ADMIN_HOST`, and `PG_ADMIN_PASSWORD`

## GitHub Actions Image Publishing

The workflow at `.github/workflows/docker-publish.yml` publishes the app image to `ghcr.io/iyouknow/10db`.

It runs on:
- pushes of tags like `v0.1.0`
- manual runs from the Actions tab

Published tags include:
- `latest`
- the git tag name, like `v0.1.0`

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
