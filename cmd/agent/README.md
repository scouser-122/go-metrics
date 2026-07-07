# cmd/agent

В данной директории будет содержаться код Агента, который скомпилируется в бинарное приложение.

Пример сборки и запуска из командной строки:

```bash
go build -ldflags "-X main.buildVersion=v1.0.0 -X main.buildDate=$(date -u '+%Y-%m-%d_%H:%M:%S') -X main.buildCommit=$(git rev-parse HEAD)" -o agent && \
./agent -config="./config.json"
```