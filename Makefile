BINARY_NAME = tetra-mess
VERSION_NUMBER ?= $(shell git describe --tags | sed -E 's#v##')

all: clean test build

clean:
	go clean
	rm -f ${BINARY_NAME}

version_number:
	@echo ${VERSION_NUMBER}

test:
	go test -v -timeout=30s ./...

build:
	go build -trimpath -buildmode=pie -mod=readonly -modcacherw -v -ldflags "-linkmode external -extldflags \"${LDFLAGS}\" -X github.com/ftl/tetra-mess/cmd.version=${VERSION_NUMBER}" -o ${BINARY_NAME}

run: build
	./${BINARY_NAME}
