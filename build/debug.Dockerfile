FROM golang:1.23.4-alpine3.21 AS builder
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
WORKDIR /app
RUN apk add --update make git
COPY . .
RUN set -x \
    && apk --no-cache add ca-certificates \
    && go install -ldflags "-s -w -extldflags '-static'" github.com/go-delve/delve/cmd/dlv@latest \
    && make build-debug
EXPOSE 2345
CMD [ "/go/bin/dlv", "--listen=:2345", "--headless=true", "--log=true", "--accept-multiclient", "--api-version=2", "exec", "/app/bin/homyaksrv", "run" ]