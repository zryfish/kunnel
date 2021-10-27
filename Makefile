VERSION=$(shell git describe --abbrev=0 --tags 2>&1)
BUILD=$(shell git rev-parse HEAD)
REPO=jeffwithlove/kunnel
TAG=${VERSION}

LDFLAGS=-ldflags "-s -w -X github.com/zryfish/kunnel/pkg/version.BuildVersion=${VERSION}"

all: server kn

server: test
	CGO_ENABLED=0 go build -trimpath ${LDFLAGS} -o bin/server cmd/server/main.go

kn: test
	CGO_ENABLED=0 go build -trimpath ${LDFLAGS} -o bin/kn cmd/kn/main.go

test: fmt vet

docker:
	docker build -t ${REPO}:${TAG} .

# Run tests
test:  fmt vet
	go test ./... -coverprofile cover.out

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...
