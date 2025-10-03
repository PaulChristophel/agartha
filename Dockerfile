# Build the application
FROM docker.io/pcm0/build-base:musl as builder-musl
ARG ENV=debug
ARG SHORT_SHA=devel
ARG GITHUB_SHA=development
ARG GITHUB_DATE='2006-01-02T15:04:05Z'
ARG GITHUB_PSEUDODATE='20060102150405'
ARG GITHUB_VERSION='0.0.0'
ARG USER_ID=1000
RUN addgroup agartha -g ${USER_ID} -S
RUN adduser -u ${USER_ID} -s /sbin/nologin -h /opt/agartha -SD -G agartha agartha

COPY go.mod go.sum Makefile ./
COPY web/package.json web/pnpm-lock.yaml ./web/
RUN ENV=${ENV} make configure
COPY . .
RUN ENV=${ENV} make build

# Export it to alpine
FROM alpine:edge AS alpine
ARG ENV=debug
ARG USER_ID=1000
COPY --from=builder-musl /etc/passwd /etc/passwd
COPY --from=builder-musl /etc/group /etc/group
COPY --from=builder-musl /opt/agartha /opt/agartha
COPY --from=builder-musl /usr/src/app/bin/${ENV}/agartha /usr/local/bin/agartha
RUN apk upgrade --update --no-cache && apk add --update --no-cache curl ca-certificates tzdata
USER ${USER_ID}:${USER_ID}
ENTRYPOINT ["/usr/local/bin/agartha"]

# Export it to scratch
FROM scratch as slim
ARG ENV=debug
ARG USER_ID=1000
# Import dependencies
# COPY --from=builder-musl /lib/ld-musl-x86_64.so.1 /lib/ld-musl-x86_64.so.1
COPY --from=builder-musl /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder-musl /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder-musl /etc/passwd /etc/passwd
COPY --from=builder-musl /etc/group /etc/group
COPY --from=builder-musl /opt/agartha /opt/agartha
COPY --from=builder-musl /usr/src/app/bin/${ENV}/agartha /usr/local/bin/agartha
USER ${USER_ID}:${USER_ID}
ENTRYPOINT ["/usr/local/bin/agartha"]