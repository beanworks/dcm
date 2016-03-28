.PHONY: test clean build

VENDOR_FLAG = GO15VENDOREXPERIMENT=1
GO_CMD = $(VENDOR_FLAG) godep go
PKG = $$(go list ./... | grep -v /vendor/)

build: bin/dcm

cross: build
	env GOOS=darwin GOARCH=amd64 $(GO_CMD) build -o bin/dcm-darwin-amd64 ./src
	env GOOS=freebsd GOARCH=amd64 $(GO_CMD) build -o bin/dcm-freebsd-amd64 ./src
	env GOOS=linux GOARCH=amd64 $(GO_CMD) build -o bin/dcm-linux-amd64 ./src
	env GOOS=windows GOARCH=amd64 $(GO_CMD) build -o bin/dcm-windows-amd64.exe ./src

test:
	$(GO_CMD) vet $(PKG)
	$(GO_CMD) test $(PKG)

vtest:
	$(GO_CMD) vet -v $(PKG)
	$(GO_CMD) test -v -cover $(PKG)

clean:
	$(GO_CMD) clean $(PKG)
	rm -f bin/dcm

cleanall: clean
	rm -f bin/dcm-*

cover:
	@echo "mode: count" > c.out
	@for pkg in $(PKG); do \
		$(GO_CMD) test -coverprofile c.out.tmp $$pkg; \
		tail -n +2 c.out.tmp >> c.out; \
	done
	$(GO_CMD) tool cover -html=c.out

coveralls:
	# go test -covermode=count -coverprofile c.out ./...
	# goveralls -service=travis-ci -coverprofile=c.out
	@echo "mode: count" > c.out
	@for pkg in $(PKG); do \
		$(GO_CMD) test -covermode=count -coverprofile c.out.tmp $$pkg; \
		tail -n +2 c.out.tmp >> c.out; \
	done
	goveralls -service=travis-ci -coverprofile=c.out

bin/dcm: src/*.go
	$(GO_CMD) build -o bin/dcm ./src
