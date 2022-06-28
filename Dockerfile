FROM golang:1.18-alpine AS builder

RUN apk add --update --no-cache \
  ca-certificates tzdata openssh git mercurial && update-ca-certificates \
  && rm -rf /var/cache/apk/*

WORKDIR /src

COPY go.mod* go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
  CGO_ENABLED=0 go install ./...

FROM alpine

RUN adduser -S -D -H -h /app appuser
USER appuser

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/* /bin/

WORKDIR /etc/twitch-subscriber-sync/
COPY --from=builder /src/settings.cfg.example settings.cfg

# TODO: Doesn't seem that the website Addr is used anywhere
ENV PORT=8080
EXPOSE $PORT

CMD ["twitchpubsub"]
