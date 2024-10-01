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
```


| flag | ENV               | usage                                     | default        | example                                                                      |
| ---- | ----------------- | ----------------------------------------- | -------------- | ---------------------------------------------------------------------------- |
| -a   | SERVER_ADDRESS    | address and port to run server            | localhost:8080 |                                                                              |
| -b   | BASE_URL          | address and port for include in shortURLs | localhost:8080 |                                                                              |
| -f   | FILE_STORAGE_PATH | path to the data storage file             | -              | ./storage.db                                                                 |
| -d   | DATABASE_DSN      | database connection string                | -              | host=127.0.0.1 user=shortener password=pass dbname=shortener sslmode=disable |

`go run ./cmd/shortener -d "host=127.0.0.1 user=shortener password=pass dbname=shortener sslmode=disable" -l debug`

`go run ./cmd/shortener -f ./storage.db -l debug`

