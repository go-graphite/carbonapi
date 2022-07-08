FROM golang:alpine as builder
RUN apk --no-cache add make gcc git cairo-dev musl-dev