.PHONY: vet
vet: 
	go vet ./... 

.PHONY: test
test:
	go test --race -v ./...

.PHONY: deps
deps:
	go mod verify
	go mod tidy
