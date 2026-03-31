# cmd/server

В данной директории будет содержаться код Сервера, который скомпилируется в бинарное приложение.

Пример сборки и запуска из командной строки:

go build -o server && ./server \
-a localhost:45057  \
-l info \
-e dev \
-i 2 \
-f="./metrics_data.json"  \
-r \
-d="postgres://postgres:123@localhost:5432/go_test?sslmode=disable"