FROM golang:alpine as builder

WORKDIR /go/src/github.com/go-graphite/carbonapi

COPY . .

RUN apk --no-cache add make gcc git cairo-dev musl-dev
RUN make && make test

FROM alpine:latest

RUN apk --no-cache add ca-certificates cairo
WORKDIR /

COPY --from=builder /go/src/github.com/go-graphite/carbonapi/carbonapi ./usr/bin/carbonapi

CMD ["carbonapi", "-config", "/etc/carbonapi.yml"]
