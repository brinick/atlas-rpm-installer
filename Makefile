# Makefile
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=atlas-rpm-install
BINARY_UNIX=$(BINARY_NAME)_unix

build: 
	$(GOBUILD) -o $(BINARY_NAME) -v -ldflags="-w -s" github.com/brinick/atlas-rpm-installer/cmd/installer
	upx $(BINARY_NAME)

test:
	$(GOTEST) -v ./...

