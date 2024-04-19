APP_NAME = gochirpy

run: build
	./bin/$(APP_NAME)

build:
	go build -o bin/$(APP_NAME) main.go
