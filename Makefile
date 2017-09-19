DIR=~/docker-alertd-dist

build:
	@rm -r $(DIR)
	@mkdir -p $(DIR)/osx $(DIR)/linux $(DIR)/windows
	@go build -o $(DIR)/osx/docker-alertd
	@GOOS=linux go build -o $(DIR)/linux/docker-alertd
	@GOOS=windows go build -o $(DIR)/windows/docker-alertd.exe

