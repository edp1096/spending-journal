.PHONY: default
default: build

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

fname := server
fext :=
ifeq ($(GOOS),Windows_NT)
	fname := server
	fext := .exe
endif


build:
	go build -trimpath -ldflags="-w -s" -o bin/$(fname)$(fext) ./cmd

clean:
	rm -rf ./bin/*
