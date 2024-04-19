APP_NAME = gochirpy

run: build clean
	./bin/$(APP_NAME)

build:
	go build -o bin/$(APP_NAME) .

clean:
	rm database.json
