FROM node:24-alpine AS web-builder
WORKDIR /src/apps/web
COPY apps/web/package.json apps/web/package-lock.json* ./
RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi
COPY apps/web/ ./
RUN npm run build

FROM golang:1.25-alpine AS go-builder
WORKDIR /src/apps/server
COPY apps/server/go.mod apps/server/go.sum* ./
RUN go mod download
COPY apps/server/ ./
COPY --from=web-builder /src/apps/web/dist ./static
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/10db-launch ./cmd/10db-launch

FROM alpine:3.22
WORKDIR /app
RUN adduser -D appuser && mkdir -p /data && chown -R appuser:appuser /data /app
USER appuser
COPY --from=go-builder /out/10db-launch /app/10db-launch
COPY --from=web-builder /src/apps/web/dist /app/static
EXPOSE 8080
ENTRYPOINT ["/app/10db-launch"]
