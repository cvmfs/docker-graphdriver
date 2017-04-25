#!/bin/bash
BINARY_PKG_URL="https://storage.googleapis.com/golang/go1.7.5.linux-amd64.tar.gz"

wget --quiet "$BINARY_PKG_URL"
tar -C /usr/local -xzf "go1.7.5.linux-amd64.tar.gz"
mkdir -p "$GOPATH/src"
