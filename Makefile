BINARY := sp13sx

.PHONY: build

build:
	go build -o $(BINARY) ./cmd/sp13sx
