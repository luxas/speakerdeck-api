all: build
build:
	go build -o bin/api ./cmd/api

run: build
	bin/api
