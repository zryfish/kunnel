
all: client server kubectl-kn

client: test
	go build -o bin/client cmd/client/main.go

server: test
	go build -o bin/server cmd/server/main.go

kubectl-kn: test
	go build -o bin/kubectl-kn cmd/kn/main.go

test: fmt vet

# Run tests
test:  fmt vet
	go test ./... -coverprofile cover.out

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...
