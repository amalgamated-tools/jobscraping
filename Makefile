# Before using, configure goproxy; see https://goproxy.githubapp.com/setup.
.PHONY: all
all: build test

.PHONY: build
build: cli

.PHONY: clean
clean:
	rm -f tools
	go clean -cache -testcache -modcache

.PHONY: cli
cli:
	go build -ldflags="-X main.Commit=$(git rev-parse HEAD)" -o bin/cli ./cmd/cli		

.PHONY: fmt
fmt:
	@echo "==> running Go format <=="
	gofmt -s -l -w .

.PHONY: vet
vet:
	@echo "==> vetting Go code <=="
	go vet ./...

.PHONY: tidy
tidy:
	@echo "==> running Go mod tidy <=="
	go mod tidy

.PHONY: sec
sec:
	@echo "==> running Go security checks <=="
	gosec -quiet ./...

.PHONY: test
test:
	@echo "==> running Go tests <=="
	go test -p 1 ./...

.PHONY: lint
lint:
	./script/lint

.PHONY: testlint
testlint: test lint
	
.PHONY: coverage
coverage:
	go test -count=1 -coverpkg=./... -covermode=atomic -coverprofile coverage.out  ./...
	cat coverage.out | grep -v "mock" > coverage.filtered.out
	mv coverage.filtered.out coverage.out
	go tool cover -func coverage.out