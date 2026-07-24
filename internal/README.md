# internal

В этой директории размещается код внутренних модулей приложения. Код внутри этого пакета недоступен для импорта в других приложениях.

Структуру дирктории `internal/` можно разбивать по логическим блокам приложения, выделяя пакеты по функциональному назначению. 
Например, `internal/agent`, `internal/server` и т.д.

Директория `internal/` является специальной в Go и обеспечивает инкапсуляцию кода на уровне модуля. Компилятор Go запрещает импорт пакетов из `internal/` за пределами родительского модуля.

Генерация proto исходников:
```bash
protoc \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  --go_opt=default_api_level=API_OPAQUE \
  internal/proto/metrics.proto 
```