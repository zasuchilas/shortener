# SHORTENER

URL shortening service

- [go-autotests](https://github.com/Yandex-Practicum/go-autotests)
- [Курс «Продвинутый Go‑разработчик»](https://practicum.yandex.ru/go-advanced/)
- [Карты курсов Go с нуля и Продвинутый Go](https://code.s3.yandex.net/go/1f_vs_2f.pdf)


Commands:  
```shell
go test -v ./...
go clean -testcache

go run ./cmd/shortener/
go build -o shortener
./shortener -a ":8033" -b ":8034"

go test -bench=. ./... -benchmem
curl -sK -v http://localhost:8080/debug/pprof/heap > profiles/base.pprof
go tool pprof -http=":9090" profiles/base.pprof
curl -sK -v http://localhost:8080/debug/pprof/heap > profiles/result.pprof
go tool pprof -http=":9090" profiles/result.pprof
go tool pprof -http=":9090" -top -diff_base=profiles/base.pprof profiles/result.pprof

goimports -local "github.com/zasuchilas/shortener" -w ./cmd/shortener/
goimports -local "github.com/zasuchilas/shortener" -w ./internal/

godoc -http=:8083 -play
# http://localhost:8083/pkg/github.com/zasuchilas/shortener/?m=all

```


| flag | ENV               | usage                                     | default        | example                                                                      |
| ---- | ----------------- | ----------------------------------------- | -------------- | ---------------------------------------------------------------------------- |
| -a   | SERVER_ADDRESS    | address and port to run server            | localhost:8080 |                                                                              |
| -b   | BASE_URL          | address and port for include in shortURLs | localhost:8080 |                                                                              |
| -f   | FILE_STORAGE_PATH | path to the data storage file             | -              | ./storage.db                                                                 |
| -d   | DATABASE_DSN      | database connection string                | -              | host=127.0.0.1 user=shortener password=pass dbname=shortener sslmode=disable |

`go run ./cmd/shortener -d "host=127.0.0.1 user=shortener password=pass dbname=shortener sslmode=disable" -l debug`

`go run ./cmd/shortener -f ./storage.db -l debug`
