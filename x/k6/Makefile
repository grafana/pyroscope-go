.PHONY: test
test:
	go test -race ./...

.PHONY: go/mod
go/mod:
	GO111MODULE=on go mod download
	GO111MODULE=on go mod tidy
