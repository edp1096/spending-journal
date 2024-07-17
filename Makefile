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

# dist:
# 	go get -d github.com/mitchellh/gox
# 	go build -mod=readonly -o ./bin/ github.com/mitchellh/gox
# 	go mod tidy
# 	go env -w GOFLAGS=-trimpath
# 	./bin/gox -mod="readonly" -output="./bin/$(fname)_{{.OS}}_{{.Arch}}$(fext)" -osarch="windows/amd64 linux/amd64 linux/arm linux/arm64 darwin/amd64 darwin/arm64" ./cmd
# 	rm ./bin/gox*

# test:
# 	go test ./... -race -cover -count=1

clean:
	rm -rf ./bin/*
