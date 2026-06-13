# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**


## Профилирование

### Статистика по выделяемой памяти (открыть в браузере)
go tool pprof -http=":9090" -seconds=30 http://localhost:6060/debug/pprof/heap

### Статистика по выделяемой памяти (сохранить в файл)
curl -o profiles/base.pprof 'http://localhost:6060/debug/pprof/heap?seconds=30'

### Результат сокращения кол-ва выделяемой памяти в функции processListRequest при вызове API GET / для получения списка метрик:

```bash
% go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof   
File: __debug_bin1380477432
Type: inuse_space
Time: 2026-06-13 13:10:26 MSK
Duration: 60.01s, Total samples = 548.84kB 
Showing nodes accounting for 1697.43kB, 309.27% of 548.84kB total
      flat  flat%   sum%        cum   cum%
    1539kB 280.41% 280.41%     1539kB 280.41%  runtime.allocm
 -902.59kB 164.45% 115.96%  -353.74kB 64.45%  compress/flate.NewWriter
  548.84kB   100% 215.96%   548.84kB   100%  compress/flate.(*compressor).initDeflate
  512.17kB 93.32% 309.27%   512.17kB 93.32%  net/textproto.MIMEHeader.Set
         0     0% 309.27%   548.84kB   100%  compress/flate.(*compressor).init
         0     0% 309.27%  -353.74kB 64.45%  compress/gzip.(*Writer).Write
         0     0% 309.27%   158.43kB 28.87%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 309.27%   158.43kB 28.87%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0% 309.27%   158.43kB 28.87%  github.com/scouser-122/go-metrics/internal/handler.(*ReadHandler).ListHandler
         0     0% 309.27%  -353.74kB 64.45%  github.com/scouser-122/go-metrics/internal/handler.(*ReadHandler).processListRequest
         0     0% 309.27%  -353.74kB 64.45%  github.com/scouser-122/go-metrics/internal/handler.(*gzipWriter).Write
         0     0% 309.27%   158.43kB 28.87%  github.com/scouser-122/go-metrics/internal/handler.GzipMiddleware.func1
         0     0% 309.27%   158.43kB 28.87%  github.com/scouser-122/go-metrics/internal/handler.RequestLogger.func1
         0     0% 309.27%  -353.74kB 64.45%  github.com/scouser-122/go-metrics/internal/model.(*LoggingResponseWriter).Write
         0     0% 309.27%   158.43kB 28.87%  net/http.(*conn).serve
         0     0% 309.27%   158.43kB 28.87%  net/http.HandlerFunc.ServeHTTP
         0     0% 309.27%   512.17kB 93.32%  net/http.Header.Set
         0     0% 309.27%   158.43kB 28.87%  net/http.serverHandler.ServeHTTP
         0     0% 309.27%     1026kB 186.94%  runtime.mcall
         0     0% 309.27%      513kB 93.47%  runtime.mstart
         0     0% 309.27%      513kB 93.47%  runtime.mstart0
         0     0% 309.27%      513kB 93.47%  runtime.mstart1
         0     0% 309.27%     1539kB 280.41%  runtime.newm
         0     0% 309.27%     1026kB 186.94%  runtime.park_m
         0     0% 309.27%     1539kB 280.41%  runtime.resetspinning
         0     0% 309.27%     1539kB 280.41%  runtime.schedule
         0     0% 309.27%     1539kB 280.41%  runtime.startm
         0     0% 309.27%     1539kB 280.41%  runtime.wakep
```