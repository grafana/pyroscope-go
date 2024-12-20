.PHONY: test
test:
	go test -race ./...

.PHONY: go/mod
go/mod:
	GO111MODULE=on go mod download
	go work sync
	GO111MODULE=on go mod tidy
