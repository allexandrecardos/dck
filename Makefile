# Makefile
BINARY_NAME=dck
VERSION?=dev
LDFLAGS=-ldflags"-X main.version=${VERSION}"

.PHONY: build install test clean dev

build build: :
	go build go build $ ${ {LDFLAGS LDFLAGS} } -o bin/ -o bin/$ ${ {BINARY_NAME BINARY_NAME} } main.go main.go

install install: : build build
	sudo cp bin/ sudo cp bin/$ ${ {BINARY_NAME BINARY_NAME} } /usr/local/bin/ /usr/local/bin/

test test: :
	go test -v ./... go test -v ./...

clean clean: :
	rm -rf bin/ rm -rf bin/

dev dev: : build build
	./bin/ ./bin/$ ${ {BINARY_NAME BINARY_NAME} }