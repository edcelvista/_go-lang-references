#!/bin/bash
GOOS=linux GOARCH=arm go build -o bin/k8swatcher-arm
GOOS=darwin GOARCH=arm64 go build -o bin/k8swatcher-darwin-arm64

# CGO_ENABLED=0 to remove C lib dependencies
# k8swatcher --action watchutil