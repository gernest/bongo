TPL_DIR:=./templates 
TPL_EMBED=bindata/tpl/tpl.go

.PHONY: all bindata

all: bindata

bindata:
	@go-bindata -pkg=tpl -o=$(TPL_EMBED) -prefix=templates/ templates/...