#!/usr/bin/env bash
set -eufo pipefail

pushd .drone

go build -o coverage ./cmd/coverage
go build -o ghcomment ./cmd/ghcomment

popd