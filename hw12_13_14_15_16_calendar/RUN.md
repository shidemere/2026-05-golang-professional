# Запуск проекта

## In-memory хранилище

По умолчанию в конфиге включено in-memory хранилище:

```bash
make run
```

Сервис запустит HTTP-сервер на `localhost:8080`.

Проверить hello endpoint:

```bash
curl http://localhost:8080/hello
```

## PostgreSQL хранилище

Запустить контейнер PostgreSQL:

```bash
make run-postgres
```

Применить миграции во временном Docker-контейнере:

```bash
make migrate
```

Если логин, пароль или имя базы отличаются от значений по умолчанию, их можно передать так:

```bash
POSTGRES_USER=postgres POSTGRES_PASSWORD=password POSTGRES_DB=backend make migrate
```

Чтобы запустить приложение с SQL-хранилищем, установите `use_in_memory: false` в `configs/config.yaml` и передайте настройки подключения через переменные окружения:

```bash
POSTGRES_USER=postgres \
POSTGRES_PASSWORD=password \
POSTGRES_DATABASE=backend \
POSTGRES_HOST=localhost \
POSTGRES_PORT=5435 \
make run
```

## Тесты

```bash
make test
```

Если локальный Go build cache доступен только для чтения, запустите тесты так:

```bash
GOCACHE=/tmp/go-build-cache go test -race ./internal/...
```

## Сборка

```bash
make build
```
