GO_VERSION_PRE20 := $(shell go version | awk '{print $$3}' | awk -F '.' '{print ($$1 == "go1" && int($$2) < 20)}')

# x/k6 has a dependency on google.golang.org/grpc, which is only compatiable
# with the last two stable versions of Go. This is a check to see if we are
# building with a version of go that is stable.
#
# See: https://github.com/grpc/grpc-go?tab=readme-ov-file#prerequisites
GO_VERSION_STABLE := $(shell scripts/check_go_stable)
ifndef GO_VERSION_STABLE
	$(error "failed to check if go version is stable")
endif

TEST_PACKAGES := ./... ./godeltaprof/compat/... ./godeltaprof/...
ifeq ($(GO_VERSION_STABLE), 1)
	TEST_PACKAGES += ./x/k6/...
endif

.PHONY: test
test:
	go test -race $(shell go list $(TEST_PACKAGES) | grep -v /example)

.PHONY: go/mod
go/mod:
	GO111MODULE=on go mod download
	go work sync
	GO111MODULE=on go mod tidy
	cd godeltaprof/compat/ && GO111MODULE=on go mod download
	cd godeltaprof/compat/ && GO111MODULE=on go mod tidy
	cd godeltaprof/ && GO111MODULE=on go mod download
	cd godeltaprof/ && GO111MODULE=on go mod tidy

ifeq ($(GO_VERSION_STABLE), 1)
	cd x/k6/ && GO111MODULE=on go mod download
	cd x/k6/ && GO111MODULE=on go mod tidy
endif

# https://github.com/grafana/pyroscope-go/issues/129
.PHONY: gotip/fix
gotip/fix:
	cd godeltaprof/compat/ && gotip get -d -v golang.org/x/tools@v0.25.0
	git --no-pager diff
	! git diff | grep toolchain
