all: build
build:
	go build -o bin/speakerdeck-api ./cmd/speakerdeck-api

run: build
	bin/speakerdeck-api
