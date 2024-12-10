# run

run:
	@go run ./cmd/shortener/
	# go run -ldflags "-X main.buildVersion=$(git describe --tags --abbrev=0) -X 'main.buildDate=$(date +'%Y.%m.%d %H:%M:%S')' -X main.buildCommit=$(git rev-parse HEAD)" ./cmd/shortener/

run_f:
	@go run ./cmd/shortener -f ./storage.db -l debug

run_pg:
	@go run ./cmd/shortener -d "host=127.0.0.1 user=shortener password=pass dbname=shortener sslmode=disable" -l debug

# lint

goimports:
	@goimports -local "github.com/zasuchilas/shortener" -w ./cmd/shortener/
	@goimports -local "github.com/zasuchilas/shortener" -w ./internal/

godoc:
	godoc -http=:8083 -play

staticcheck_install:
	go install honnef.co/go/tools/cmd/staticcheck@latest

staticcheck_run:
	staticcheck ./...

# test

test:
	@go test -v ./...

clean_test_cache:
	@go clean -testcache

test_coverage:
	/usr/local/go/bin/go test -json ./... -covermode=atomic -coverprofile /home/zasuchilas/.cache/JetBrains/GoLand2021.2/coverage/shortener$go_test_shortener.out


# build

build_shortener:
	@cd ./cmd/shortener && go build -o shortener
	# cd ./cmd/shortener && go build -ldflags "-X main.buildVersion=$(git describe --tags --abbrev=0) -X 'main.buildDate=$(date +'%Y.%m.%d %H:%M:%S')' -X main.buildCommit=$(git rev-parse HEAD)" -o shortener

build_staticlint:
	@cd ./cmd/staticlint && go build -o staticlint

staticlint_help:
	@./cmd/staticlint/staticlint

staticlint_run:
	@./cmd/staticlint/staticlint ./...
