{
  _images+:: {
    argoCli: 'us.gcr.io/kubernetes-dev/drone/plugins/argo-cli',
    go: 'golang:1.17',
    goLint: 'golangci/golangci-lint:v1.45',
    dind: 'docker:dind',
    testRunner: 'us.gcr.io/kubernetes-dev/carbonapi/test-runner:latest',
  },
}