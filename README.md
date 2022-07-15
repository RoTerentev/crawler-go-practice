краулер, который в N потоков, но не более 15 страниц в секунду, обкачивает https://www.ietf.org/ с заданной страницы и вглубь по всем ссылкам на другие RFC ("[RFC 1035]").

## Help (dev)
```sh
go run cmd/main.go --help
```

## Run (dev)
```sh
go run cmd/main.go -n 2 -r 1
```