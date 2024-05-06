FROM golang:alpine AS builder

WORKDIR /usr/src/app
RUN apk upgrade --update --no-cache
RUN apk add --update --no-cache make tzdata ca-certificates nodejs-current yarn npm && npm install -g pnpm
COPY go.mod go.sum Makefile ./
RUN make go-configure
COPY . .
RUN make web-configure
# RUN make upgrade
RUN ENV=PRODUCTION make build
ARG USER_ID=1000
RUN addgroup agartha -g ${USER_ID} -S
RUN adduser -u ${USER_ID} -s /sbin/nologin -h /opt/agartha -SD -G agartha agartha

FROM alpine:edge AS alpine
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /opt/agartha /opt/agartha
COPY --from=builder /usr/src/app/bin/release /usr/local/bin
RUN apk upgrade --update --no-cache && apk add --update --no-cache curl git git-lfs ca-certificates openssh-client
USER ${USER_ID}:${USER_ID}
ENTRYPOINT ["/usr/local/bin/agartha"]

FROM busybox:latest as busybox
# Import dependencies
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /opt/agartha /opt/agartha
# Import the application
COPY --from=builder /usr/src/app/bin/release /usr/local/bin
# Use an unprivileged user
USER ${USER_ID}:${USER_ID}
# Run
ENTRYPOINT ["/usr/local/bin/agartha"]

FROM scratch as slim
# Import dependencies
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /opt/agartha /opt/agartha
# Import the application
COPY --from=builder /usr/src/app/bin/release /usr/local/bin
# Use an unprivileged user
USER ${USER_ID}:${USER_ID}
# Run
ENTRYPOINT ["/usr/local/bin/agartha"]