APP_NAME = gochirpy

run: build
	./bin/$(APP_NAME) --debug

build:
	go build -o bin/$(APP_NAME) .

