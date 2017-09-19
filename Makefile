DIR=~/docker-alertd-dist

build:
	@rm -r $(DIR)
	@mkdir -p $(DIR)/osx $(DIR)/linux $(DIR)/windows
	@go build -o ~/dist/osx/docker-alertd
	@GOOS=linux go build -o ~/dist/linux/docker-alertd
	@GOOS=windows go build -o ~/dist/windows/docker-alertd.exe

