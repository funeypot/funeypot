NAME=sshless
IMAGE_NAME=registry.aliyuncs.com/wolfogre/$(NAME)
VERSION=$(shell git describe --tags --dirty --broken --always)

tidy:
	go mod tidy -v

check: build
	git diff HEAD --quiet || exit 1

build: tidy
	mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -v -o bin/linux/$(NAME)

image: check build
	docker build -t $(IMAGE_NAME):$(VERSION) .

push:
	docker push $(IMAGE_NAME):$(VERSION)

clean:
	rm -rf bin

all: image push clean
	git push origin main --tags

run:
	go build -v -o bin/darwin/$(NAME)
	bin/darwin/$(NAME)

bump:
	rm -f go.mod
	go mod init $(NAME)
	go mod tidy -v
