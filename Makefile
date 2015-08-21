cmd_dir:=cmd/bongo
.PHONY: all
all:
	@go vet
	@golint
	@cd $(cmd_dir)&&go build

clean:
	@cd $(cmd_dir)&&go clean

deps:
	@go get github.com/mitchellh/gox

dist:
	gox -output="dist/{{.Dir}}_{{.OS}}_{{.Arch}}" ./cmd/bongo 