#!/usr/bin/env bash
set -e
echo "building linux binary"
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o bfs-linux-amd64
echo "building android binary"
GOOS=android GOARCH=arm64 go build -ldflags "-s -w" -trimpath -o bfs-android-arm64
echo "building wendos binary"
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o bfs-windows-amd64.exe