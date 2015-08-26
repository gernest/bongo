cmd_dir:=cmd/bongo
.PHONY: all dist
ifeq "$(origin APP_VER)" "undefined"
VERSION=0.1
endif
all: test
	@go vet
	@golint
	@cd $(cmd_dir)&&go build

clean:
	@cd $(cmd_dir)&&go clean

deps:
	@go get github.com/mitchellh/gox

dist:
	-@rm -r dist
	@gox -output="dist/{{.Dir}}.v$(VERSION)_{{.OS}}_{{.Arch}}/{{.Dir}}" ./cmd/bongo

test:
	@go test 

install:
	@cd $(cmd_dir)&&go install