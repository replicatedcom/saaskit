.PHONY: vet
vet: 
	go vet ./pkg/... 

.PHONY: test
test:
	go test --race -v ./pkg/...

.PHONY: deps
deps:
	go mod verify
	go mod tidy
