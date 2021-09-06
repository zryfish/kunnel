VERSION=$(shell git describe --abbrev=0 --tags 2>1)
BUILD=$(shell git rev-parse HEAD)
REPO=jeffwithlove/kunnel
TAG=${VERSION:-latest}

LDFLAGS=-ldflags "-s -w -X github.com/zryfish/kunnel/pkg/version.BuildVersion=${VERSION}"

all: client server kubectl-kn

client: test
	CGO_ENABLED=0 go build -trimpath ${LDFLAGS} -o bin/client cmd/client/main.go

server: test
	CGO_ENABLED=0 go build -trimpath ${LDFLAGS} -o bin/server cmd/server/main.go

kubectl-kn: test
	CGO_ENABLED=0 go build -trimpath ${LDFLAGS} -o bin/kubectl-kn cmd/kn/main.go

test: fmt vet

docker:
	@docker build -t ${REPO}:${TAG} .

# Run tests
test:  fmt vet
	go test ./... -coverprofile cover.out

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...
