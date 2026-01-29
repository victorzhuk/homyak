# Stage 1: Build frontend
FROM node:24.13.0-slim AS nodejs
WORKDIR /web
COPY web .
RUN yarn install
RUN yarn build

# Stage 2: Build Go binary
FROM golang:1.25.6-bookworm AS builder
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends make git && rm -rf /var/lib/apt/lists/*
COPY . .
COPY --from=nodejs /web/dist /app/web/dist
RUN make build

# Stage 3: Runtime - distroless for minimal attack surface
FROM gcr.io/distroless/static-debian13:nonroot
WORKDIR /app
COPY --from=builder /app/bin/homyaksrv .
# Note: distroless:nonroot already runs as uid:nonroot (65532)
EXPOSE 8080/tcp
CMD ["/app/homyaksrv", "run"]