build:
	gofmt -s -w ./
	go get
	env GOOS=linux go build -ldflags="-s -w" -x -o bin/bugcaster ./

.PHONY: clean
clean:
	rm -rf ./bin ./vendor Gopkg.lock

.PHONY: deploy
deploy: clean build
	sls deploy --verbose
