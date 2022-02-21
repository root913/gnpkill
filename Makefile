BINARY_NAME=gnpkill

build_darwin:
	GOARCH=amd64 GOOS=darwin go build -o bin/${BINARY_NAME}-darwin

build_linux:
	GOARCH=amd64 GOOS=linux go build -o bin/${BINARY_NAME}-linux

build_windows:
	GOARCH=amd64 GOOS=windows go build -o bin/${BINARY_NAME}-windows.exe

build: build_darwin build_linux build_windows

clean:
	go clean
	rm bin/${BINARY_NAME}-darwin
	rm bin/${BINARY_NAME}-linux
	rm bin/${BINARY_NAME}-windows.exe