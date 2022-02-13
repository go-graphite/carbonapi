FROM golang:alpine as builder

WORKDIR /go/src/github.com/go-graphite/carbonapi

COPY . .

RUN apk --no-cache add make gcc git cairo-dev musl-dev
RUN make && make test

# If you have "Operation not permitted" errors, please refer to https://wiki.alpinelinux.org/wiki/Release_Notes_for_Alpine_3.14.0#faccessat2
# TLDR; Either update docker/moby and libseccomp or switch to alpine:3.13 (builder needs to be switched to 1.16-alpine3.13 as well).
# See https://github.com/go-graphite/carbonapi/issues/639#issuecomment-896570456 for detailed information
FROM alpine:latest

RUN apk --no-cache add ca-certificates cairo
WORKDIR /

COPY --from=builder /go/src/github.com/go-graphite/carbonapi/carbonapi ./usr/bin/carbonapi

CMD ["carbonapi", "-config", "/etc/carbonapi.yml"]
