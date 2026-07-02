# cmd/server

В данной директории будет содержаться код Сервера, который скомпилируется в бинарное приложение.

Пример сборки и запуска из командной строки:

```bash
go build -ldflags "-X main.buildVersion=v1.0.0 -X main.buildDate=$(date -u '+%Y-%m-%d_%H:%M:%S') -X main.buildCommit=$(git rev-parse HEAD)" -o server && ./server \
-a localhost:45057  \
-l info \
-e dev \
-i 2 \
-d="postgres://postgres:123@localhost:5432/metrics_db?sslmode=disable" \
-k="secret_key" \
-f="./metrics_data.json"  \
-r
```
