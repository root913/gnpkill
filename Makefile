BINARY_NAME=gnpkill

ifndef VERSION
	VERSION ?= $(shell git describe --tags --always --abbrev=0 --match='v[0-9]*.[0-9]*.[0-9]*' 2> /dev/null | sed 's/^.//')
endif

LDFLAGS += -s -w -extldflags "-static" -X "main.version=$(VERSION)"

build_darwin:
	GOARCH=amd64 GOOS=darwin GO111MODULE=on go build -ldflags '$(LDFLAGS)' -o bin/${BINARY_NAME}-darwin

build_linux:
	GOARCH=amd64 GOOS=linux GO111MODULE=on go build -ldflags '$(LDFLAGS)' -o bin/${BINARY_NAME}-linux

build_windows:
	GOARCH=amd64 GOOS=windows GO111MODULE=on go build -ldflags '$(LDFLAGS)' -o bin/${BINARY_NAME}-windows.exe

build: build_darwin build_linux build_windows

clean:
	go clean
	rm bin/${BINARY_NAME}-darwin
	rm bin/${BINARY_NAME}-linux
	rm bin/${BINARY_NAME}-windows.exe