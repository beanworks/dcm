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

vtest:
	go vet -v ./...
	go test -v -cover ./...

clean:
	go clean ./...
	rm -f bin/dcm

cleanall: clean
	rm -f bin/dcm-*

cover:
	go test -coverprofile c.out ./...
	go tool cover -html=c.out

coveralls:
	go test -covermode=count -coverprofile c.out ./...
	goveralls -service=travis-ci -coverprofile=c.out

bin/dcm: src/*.go
	go get ./...
	go build -o bin/dcm ./src
