# cmd/server

В данной директории будет содержаться код Сервера, который скомпилируется в бинарное приложение.

Пример сборки и запуска из командной строки:

go build -o server && ./server \
-a localhost:45057  \
-l info \
-e dev \
-i 2 \
-d="postgres://postgres:123@localhost:5432/metrics_db?sslmode=disable" \
-f="./metrics_data.json"  \
-r
