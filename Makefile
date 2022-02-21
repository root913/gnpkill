BINARY_NAME=gnpkill

build:
	GOARCH=amd64 GOOS=darwin go build -o bin/${BINARY_NAME}-darwin
	GOARCH=amd64 GOOS=linux go build -o bin/${BINARY_NAME}-linux
	GOARCH=amd64 GOOS=windows go build -o bin/${BINARY_NAME}-windows.exe

run:
	./${BINARY_NAME}

build_and_run: build run

clean:
	go clean
	rm ${BINARY_NAME}-darwin
	rm ${BINARY_NAME}-linux
	rm ${BINARY_NAME}-windows.exe