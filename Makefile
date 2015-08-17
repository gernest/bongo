TPL_DIR			:=./templates 
TPL_EMBED		:=bindata/tpl/tpl.go
BOWER			:=bower_components
STATIC_FILES 	:=$(BOWER)/primer-css/css/primer.css $(BOWER)/primer-markdown/dist/user-content.min.css
STATIC_EMBED		:=bindata/static/static.go

.PHONY: all bindata

all: bindata

bindata:
	@go-bindata -pkg=tpl -o=$(TPL_EMBED) -prefix=templates/ templates/...
	@go-bindata -pkg=static -o=$(STATIC_EMBED) $(STATIC_FILES)