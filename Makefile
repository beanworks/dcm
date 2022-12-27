.PHONY: test clean build

PKG = $$(go list ./... | grep -v /vendor/)

build: bin/dcm

cross: build
	env GOOS=darwin GOARCH=amd64 go build -o bin/dcm-darwin-amd64 ./src
	env GOOS=darwin GOARCH=arm64 go build -o bin/dcm-linux-amd64 ./src
	env GOOS=freebsd GOARCH=amd64 go build -o bin/dcm-freebsd-amd64 ./src
	env GOOS=linux GOARCH=amd64 go build -o bin/dcm-linux-amd64 ./src
	env GOOS=windows GOARCH=amd64 go build -o bin/dcm-windows-amd64.exe ./src

test:
	go vet $(PKG)
	go test $(PKG)

vtest:
	go vet -v $(PKG)
	go test -v -cover $(PKG)

clean:
	go clean $(PKG)
	rm -f bin/dcm

cleanall: clean
	rm -f bin/dcm-*

cover:
	@echo "mode: count" > c.out
	@for pkg in $(PKG); do \
		go test -coverprofile c.out.tmp $$pkg; \
		tail -n +2 c.out.tmp >> c.out; \
	done
	go tool cover -html=c.out

coveralls:
	# go test -covermode=count -coverprofile c.out ./...
	# goveralls -service=travis-ci -coverprofile=c.out
	@echo "mode: count" > c.out
	@for pkg in $(PKG); do \
		go test -covermode=count -coverprofile c.out.tmp $$pkg; \
		tail -n +2 c.out.tmp >> c.out; \
	done
	goveralls -service=travis-ci -coverprofile=c.out

bin/dcm: src/*.go
	go build -o bin/dcm ./src
