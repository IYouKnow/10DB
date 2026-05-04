FROM node:24-alpine AS web-builder
WORKDIR /src/apps/web
COPY apps/web/package.json apps/web/package-lock.json* ./
RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi
COPY apps/web/ ./
RUN npm run build

FROM golang:1.25-alpine AS server-builder
WORKDIR /src/apps/server
COPY apps/server/go.mod apps/server/go.sum ./
RUN go mod download
COPY apps/server/ ./
COPY --from=web-builder /src/apps/web/dist ./static
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/10db-launch .

FROM alpine:3.22
WORKDIR /app
RUN apk add --no-cache ca-certificates && \
    adduser -D -u 10001 appuser && \
    mkdir -p /app/static /app/migrations /data && \
    chown -R appuser:appuser /app /data
COPY --from=server-builder /out/10db-launch /app/10db-launch
COPY --from=server-builder /src/apps/server/migrations /app/migrations
COPY --from=web-builder /src/apps/web/dist /app/static
USER appuser
EXPOSE 8080
ENTRYPOINT ["/app/10db-launch"]
