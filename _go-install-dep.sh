#!/bin/bash
# go get k8s.io/client-go@latest
# go get k8s.io/apimachinery@latest
# go get k8s.io/kubectl@latest
# export GO111MODULE=on # To ensure Go uses only dependencies in your module (instead of system-wide ones), use:
# go mod tidy # This will clean up unused dependencies and fetch missing ones.
# go list -m all
# go mod vendor
# GOPRIVATE # The GOPRIVATE environment variable can be used to specify which repositories are private.
# export GOFLAGS="" # This forces Go to fetch dependencies from the module cache instead of vendor/.
# go list -m all | grep k8s.io/client-go

# CLEAN
go clean -modcache
go mod tidy
go mod vendor

# REINIT
rm -rf vendor
go mod vendor