.PHONY: test clean build

build: bin/dcm

test:
	go vet ./...
	go test ./...

clean:
	go clean ./...
	rm -f bin/dcm

bin/dcm: *.go
	go get ./...
	go build -o bin/dcm
