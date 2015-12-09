.PHONY: test clean build

build: bin/dcm

cross: build
	env GOOS=darwin GOARCH=amd64 go build -o bin/dcm-darwin-amd64 ./src
	env GOOS=freebsd GOARCH=amd64 go build -o bin/dcm-freebsd-amd64 ./src
	env GOOS=linux GOARCH=amd64 go build -o bin/dcm-linux-amd64 ./src
	env GOOS=windows GOARCH=amd64 go build -o bin/dcm-windows-amd64.exe ./src

test:
	go vet ./...
	go test ./...

clean:
	go clean ./...
	rm -f bin/dcm

bin/dcm: src/*.go
	go get ./...
	go build -o bin/dcm ./src
