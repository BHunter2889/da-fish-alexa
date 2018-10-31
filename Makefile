build:
	dep ensure -v
	env GOOS=linux go build -ldflags="-s -w" -x -o bin/da-fish-alexa ./

.PHONY: clean
clean:
	rm -rf ./bin ./vendor Gopkg.lock

.PHONY: deploy
deploy: clean build
	sls deploy --verbose
