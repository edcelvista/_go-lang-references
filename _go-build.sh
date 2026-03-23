#!/bin/bash
GOOS=linux GOARCH=amd64 go build -o app-linux
go mod vendor # By default, Go caches dependencies globally. However, you can vendor them inside your project: