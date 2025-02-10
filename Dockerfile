FROM golang:alpine AS builder
ARG TARGETARCH

RUN apk --no-cache add --update make gcc git musl-dev

USER nobody:nogroup
WORKDIR /usr/local/src/carbonapi
COPY --chown=nobody:nogroup . .
RUN --network=none make clean
RUN --mount=type=cache,id=go-cache,target=/.cache,sharing=locked,uid=65534,gid=65534 make nocairo
RUN --mount=type=cache,id=go-cache,target=/.cache,sharing=locked,uid=65534,gid=65534 <<EOT
if [ "${TARGETARCH:-unknown}" = "amd64" ]; then
  make test_nocairo
else
  make test_nocairo_norace
fi
EOT

# If you have "Operation not permitted" errors, please refer to https://wiki.alpinelinux.org/wiki/Release_Notes_for_Alpine_3.14.0#faccessat2
# TLDR; Either update docker/moby and libseccomp or switch to alpine:3.13 (builder needs to be switched to 1.16-alpine3.13 as well).
# See https://github.com/go-graphite/carbonapi/issues/639#issuecomment-896570456 for detailed information
FROM alpine:latest

RUN addgroup -S carbon && \
  adduser -S carbon -G carbon && \
  apk --no-cache add --update ca-certificates

COPY --chown=0:0 --from=builder /usr/local/src/carbonapi/carbonapi /usr/sbin/carbonapi
WORKDIR /etc/carbonapi
COPY --chown=0:0 --from=builder /usr/local/src/carbonapi/cmd/carbonapi/carbonapi.docker.yaml carbonapi.yaml

WORKDIR /
USER carbon
ENTRYPOINT ["/usr/sbin/carbonapi"]
CMD ["-config", "/etc/carbonapi/carbonapi.yaml"]

EXPOSE 8080
VOLUME /etc/carbonapi