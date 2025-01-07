FROM node:22.12-alpine3.21 AS nodejs
WORKDIR /web
COPY web .
RUN yarn install
RUN yarn build

FROM golang:1.23.4-alpine3.21 AS builder
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
WORKDIR /app
RUN apk add --update make git
COPY . .
COPY --from=nodejs /web/dist /app/web/dist
RUN make build

FROM alpine:3.21
ENV USER=zhuk
ENV GROUPNAME=$USER
WORKDIR /app
RUN addgroup \
  "$GROUPNAME" \
&& adduser \
  --disabled-password \
  --gecos "" \
  --home "$(pwd)" \
  --ingroup "$GROUPNAME" \
  --no-create-home \
  $USER
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/bin/homyaksrv .
RUN chown -R $USER:$GROUPNAME /app
USER $USER
CMD ["/app/homyaksrv", "run"]