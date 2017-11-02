DIR=./docker-alertd-dist

all: build docker

build:
	@echo "building distributions"
	@rm -r $(DIR)
	@mkdir -p $(DIR)/osx $(DIR)/linux $(DIR)/windows
	@go build -o $(DIR)/osx/docker-alertd
	@GOOS=linux go build -o $(DIR)/linux/docker-alertd
	@GOOS=windows go build -o $(DIR)/windows/docker-alertd.exe

docker:
	@echo "building docker image"
	@docker build -t deltaskelta/docker-alertd .
	@echo "pushing to docker registry"
	@docker push deltaskelta/docker-alertd
