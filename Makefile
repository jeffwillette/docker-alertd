
test:
	@docker pull alpine:latest
	@go test ./cmd
