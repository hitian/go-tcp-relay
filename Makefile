.PHONY: clean build

clean: 
	rm -rf bin/*
	
build:
	go build -ldflags="-s -w" -o bin/go-relay

build-all:
	bash ./build.sh
	
