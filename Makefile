cmd_dir:=cmd/bongo
.PHONY: all
all:
	@go vet
	@golint
	@cd $(cmd_dir)&&go build

clean:
	@cd $(cmd_dir)&&go clean
