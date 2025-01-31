# run

run:
	@go run ./cmd/shortener/
	# go run -ldflags "-X main.buildVersion=$(git describe --tags --abbrev=0) -X 'main.buildDate=$(date +'%Y.%m.%d %H:%M:%S')' -X main.buildCommit=$(git rev-parse HEAD)" ./cmd/shortener/

run_f:
	@go run ./cmd/shortener -f ./storage.db -l debug

run_pg:
	@go run ./cmd/shortener -d "host=127.0.0.1 user=shortener password=pass dbname=shortener sslmode=disable" -l debug

run_pg_tls:
	@go run ./cmd/shortener -d "host=127.0.0.1 user=shortener password=pass dbname=shortener sslmode=disable" -l debug -s

run_pg_config:
	@go run ./cmd/shortener -d "host=127.0.0.1 user=shortener password=pass dbname=shortener sslmode=disable" -c config.json

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

# gRPC

LOCAL_BIN:=$(CURDIR)/bin

install-deps:
	GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	GOBIN=$(LOCAL_BIN) go install -mod=mod google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

get-deps:
	go get -u google.golang.org/protobuf/cmd/protoc-gen-go
	go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc

vendor-proto:
		@if [ ! -d vendor.protogen/google ]; then \
			git clone https://github.com/googleapis/googleapis vendor.protogen/googleapis &&\
			mkdir -p  vendor.protogen/google/ &&\
			mv vendor.protogen/googleapis/google/api vendor.protogen/google &&\
			rm -rf vendor.protogen/googleapis ;\
		fi

generate:
	make generate-shortener-api

generate-shortener-api:
	mkdir -p pkg/shortenergrpcv1
	protoc --proto_path api/shortenergrpcv1 --proto_path vendor.protogen \
	--go_out=pkg/shortenergrpcv1 --go_opt=paths=source_relative \
	--plugin=protoc-gen-go=bin/protoc-gen-go \
	--go-grpc_out=pkg/shortenergrpcv1 --go-grpc_opt=paths=source_relative \
	--plugin=protoc-gen-go-grpc=bin/protoc-gen-go-grpc \
	api/shortenergrpcv1/shortener.proto
