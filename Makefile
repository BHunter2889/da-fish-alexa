build:
	dep ensure -v
	dep ensure -update -v github.com/aws/aws-xray-sdk-go
	go get
	env GOOS=linux go build -ldflags="-s -w" -x -o bin/da-fish-alexa ./

.PHONY: clean
clean:
	rm -rf ./bin ./vendor Gopkg.lock

.PHONY: deploy
deploy: clean build
	sls deploy --verbose
