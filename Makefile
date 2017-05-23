.PHONY: all

default: binary

dependencies:
	glide install

binary:
	go build

test-unit:
	go test -v -cover -coverprofile=cover.out "github.com/ldez/rebaese" ;\
    go test -v -cover -coverprofile=cover.out "github.com/ldez/rebaese/core" ;\
    go test -v -cover -coverprofile=cover.out "github.com/ldez/rebaese/gh" ;\
    go test -v -cover -coverprofile=cover.out "github.com/ldez/rebaese/git"
