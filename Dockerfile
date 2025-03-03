FROM caddy:builder-alpine AS builder

RUN xcaddy build --with github.com/simongregorebner/gitea-pages@v0.0.5


FROM alpine

COPY --from=builder /usr/bin/caddy /usr/bin/caddy

CMD ["/usr/bin/caddy", "run", "--config", "/etc/caddy/Caddyfile", "--adapter", "caddyfile"]