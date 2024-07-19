.PHONY: default
default: build

fname := server
fext :=
ifeq ($(OS),Windows_NT)
	fname := server
	fext := .exe
endif


build:
	go build -trimpath -ldflags="-w -s" -o bin/$(fname)$(fext) ./cmd

clean:
	rm -rf ./bin/*
