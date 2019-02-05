FROM golang:alpine as builder

WORKDIR /go/src/github.com/go-graphite/carbonapi

COPY . .

RUN apk --no-cache add make gcc git
RUN make nocairo

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /

COPY --from=builder /go/src/github.com/go-graphite/carbonapi/carbonapi ./usr/bin/carbonapi

CMD ["carbonapi", "-config", "/etc/carbonapi.yml"]
