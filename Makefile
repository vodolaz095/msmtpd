deps:
	go mod download
	go mod verify
	go mod tidy

test: check

lint:
	gofmt  -w=true -s=true -l=true ./
	golint ./...
	go vet ./...

check: lint
	go test -v ./...

clean:
	go clean
