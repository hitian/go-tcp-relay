#!/bin/bash

upx="upx -9"

#macOS
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o bin/go-relay_darwin_amd64

#linux amd64
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/go-relay_linux_amd64

#linux arm64
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o bin/go-relay_linux_arm64

#linux arm32v7
GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-s -w" -o bin/go-relay_linux_arm32v7

# MIPS
MIPSS=(mips mipsle)
for v in ${MIPSS[@]}; do
	GOOS=linux GOARCH=$v go build -ldflags="-s -w" -o bin/go-relay_linux_$v
	GOOS=linux GOARCH=$v GOMIPS=softfloat go build -ldflags="-s -w" -o bin/go-relay_linux_${v}_sf
done

upx -9 bin/* 
