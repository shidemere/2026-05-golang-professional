# Запуск проекта

Основной порядок: подготовить инструменты → запустить PostgreSQL → применить миграции → настроить хранилище → запустить календарь → проверить API через Postman.

## 1. Подготовить окружение

Нужны:

- Go 1.26.0 или новее, совместимый с `go.mod`.
- Docker с запущенным Docker Engine (например, Docker Desktop).
- `make` и Git для команд сборки.
- Postman с поддержкой gRPC-запросов.

Все команды выполняй из корня проекта, где находятся `Makefile` и `go.mod`. Шаги 2–5 выполняй в одном терминале: переменные окружения должны сохраниться до запуска приложения.

Проверь инструменты и скачай Go-зависимости:

```bash
go version
docker info
make --version
git --version
go mod download
```

Если `docker info` сообщает, что не может подключиться к Docker daemon, запусти Docker Desktop и дождись готовности движка.

## 2. Настроить и запустить PostgreSQL

Параметры тестового контейнера, создаваемого командами ниже, соответствуют значениям по умолчанию в `Makefile`:

| Параметр | Значение |
| --- | --- |
| Образ | `postgres:16` |
| Контейнер | `postgres-calendar` |
| Пользователь | `postgres` |
| Пароль | `password` |
| База данных | `backend` |
| Адрес для приложения и клиента БД | `localhost:5435` |
| Порт внутри контейнера | `5432` |
| Docker volume с данными | `postgres-data` |

Задай параметры явно:

```bash
export POSTGRES_USER=postgres
export POSTGRES_PASSWORD=password
export POSTGRES_DB=backend
export POSTGRES_DATABASE=backend
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5435
export POSTGRES_SSLMODE=disable
```

`POSTGRES_DB` используется Docker и миграциями, а `POSTGRES_DATABASE` — приложением. Для них указывается одна и та же база.

При первом запуске создай контейнер:

```bash
make run-postgres
```

Если контейнер `postgres-calendar` уже существует, вместо создания запусти его:

```bash
docker start postgres-calendar
```

Проверь готовность PostgreSQL:

```bash
docker exec postgres-calendar pg_isready -U postgres -d backend
```

Дождись ответа `accepting connections`. Если база ещё запускается, повтори проверку. Затем проверь подключение с паролем по TCP внутри контейнера:

```bash
docker exec -e PGPASSWORD=password postgres-calendar psql -h 127.0.0.1 -U postgres -d backend -c 'SELECT current_user, current_database();'
```

Результат должен содержать пользователя `postgres` и базу `backend`.

Если volume `postgres-data` был инициализирован ранее с другими учётными данными, новые значения `POSTGRES_USER` и `POSTGRES_PASSWORD` не изменят существующих пользователей. В этом случае используй исходные учётные данные и передай их через переменные окружения перед миграциями и запуском приложения.

## 3. Применить миграции

После готовности PostgreSQL выполни:

```bash
make migrate
```

Команда запускает временный контейнер с Go, устанавливает Goose и применяет миграции из `migrations/`. При первом запуске нужны интернет и время на загрузку образа и зависимостей. Локально устанавливать Goose не требуется.

Миграции подключаются через сеть контейнера `postgres-calendar` к `127.0.0.1:5432`. Приложение на компьютере использует опубликованный порт `localhost:5435`.

Проверь, что таблица событий создана:

```bash
docker exec postgres-calendar psql -U postgres -d backend -c '\dt'
```

В списке должна быть таблица `events`. При повторном запуске `make migrate` уже применённые миграции повторно не выполняются.

## 4. Выбрать PostgreSQL в конфигурации

В `configs/config.yaml` измени:

```yaml
use_in_memory: false
```

По умолчанию там `true`: в этом режиме приложение хранит события в памяти и не использует PostgreSQL.

Адреса серверов в этом же файле:

```yaml
http_server:
  host: localhost
  port: 8080

grpc_server:
  port: 9090
```

Логин и пароль передаются через переменные окружения из шага 2. Приложение требует `POSTGRES_USER` и `POSTGRES_PASSWORD` при работе с PostgreSQL, даже если они указаны в YAML.

## 5. Запустить календарь

В том же терминале выполни:

```bash
make run
```

Команда собирает `bin/calendar` и запускает его с `configs/config.yaml`. Один процесс обслуживает оба API:

- HTTP: `http://localhost:8080`.
- gRPC: `localhost:9090`, без TLS.

Оставь этот терминал работающим. Для следующих команд открой второй терминал в корне проекта.

RabbitMQ для текущего процесса календаря и проверки HTTP/gRPC API не нужен: `cmd/calendar` к нему не подключается.

## 6. Проверить основной функционал

Сначала проверь доступность HTTP:

```bash
curl http://localhost:8080/hello
```

Ожидаемый ответ: `Hello from server!`.

### HTTP

Импортируй `postman/calendar.postman_collection.json` в Postman. Проверь переменную `baseUrl=http://localhost:8080` и запусти **Run collection**: последовательно, одну итерацию. Коллекция содержит 9 запросов и автоматически сохраняет UUID пользователя и ID события.

После полного прогона событие удаляется. Если прогон прерван до удаления, выполни `08 Delete event` с сохранённым `eventId` перед новым запуском. Подробности: [postman/README.md](postman/README.md).

### gRPC

1. Создай запрос **New → gRPC**.
2. Укажи адрес `localhost:9090`, выключи TLS (значок замка).
3. В **Service definition → Import a .proto file** выбери [api/grpc/EventService.proto](api/grpc/EventService.proto).
4. Если Postman не находит `google/protobuf/timestamp.proto` или `google/protobuf/empty.proto`, добавь в **Import paths** каталог с папкой `google/protobuf` из установленного protobuf (обычно `/usr/include` на Linux).
5. Выбирай методы сервиса `event.EventService`, вставляй соответствующее JSON-тело ниже в **Message** и нажимай **Invoke**.

Выполни методы в порядке ниже. После `CreateEvent` скопируй `id` из ответа и замени `{{grpcEventId}}` в `UpdateEvent` и `DeleteEvent` на этот ID либо задай переменную Postman `grpcEventId`. UUID пользователя уже указан в телах.

Примеры используют 18 октября 2026 года, UTC: создание — 10:00–11:00, обновление — 12:00–13:00. Timestamp передаётся объектом с `seconds` и `nanos`.

#### CreateEvent

```json
{
  "title": "Postman gRPC meeting",
  "scheduled_at": {
    "seconds": "1792317600",
    "nanos": 0
  },
  "finished_at": {
    "seconds": "1792321200",
    "nanos": 0
  },
  "description": "Created by gRPC collection",
  "user_id": "11111111-1111-4111-8111-111111111111",
  "notify_before_seconds": "60"
}
```

#### ListByDay

```json
{
  "day": {
    "seconds": "1792281600",
    "nanos": 0
  }
}
```

#### ListByWeek

```json
{
  "week": {
    "seconds": "1792281600",
    "nanos": 0
  }
}
```

#### ListByMonth

```json
{
  "month": {
    "seconds": "1792281600",
    "nanos": 0
  }
}
```

#### UpdateEvent

```json
{
  "title": "Updated gRPC meeting",
  "scheduled_at": {
    "seconds": "1792324800",
    "nanos": 0
  },
  "finished_at": {
    "seconds": "1792328400",
    "nanos": 0
  },
  "description": "Created by gRPC collection",
  "user_id": "11111111-1111-4111-8111-111111111111",
  "notify_before_seconds": "60",
  "id": "{{grpcEventId}}"
}
```

После обновления повтори `ListByDay` с тем же телом и проверь изменённое название события. Затем удали событие:

#### DeleteEvent

```json
{
  "id": "{{grpcEventId}}"
}
```

После удаления повтори `ListByDay` и проверь, что созданного ID нет в списке. При прерванной проверке удали созданное событие через `DeleteEvent` перед новым запуском.

## 7. Остановить и запустить повторно

Останови календарь через `Ctrl+C` в терминале с `make run`. Затем останови базу:

```bash
docker stop postgres-calendar
```

Данные PostgreSQL сохраняются в volume `postgres-data`.

При следующем запуске повтори экспорт переменных из шага 2, затем выполни по порядку:

```bash
docker start postgres-calendar
docker exec postgres-calendar pg_isready -U postgres -d backend
make migrate
make run
```

Переходи к миграциям только после ответа `accepting connections`. Повторно создавать контейнер не нужно.

## Альтернатива: запуск без PostgreSQL

Для быстрого запуска установи в `configs/config.yaml`:

```yaml
use_in_memory: true
```

Затем выполни:

```bash
make run
```

Docker, миграции и переменные PostgreSQL в этом режиме не нужны. Оба API и запросы Postman работают так же, но события теряются при перезапуске процесса.

## Дополнительные команды

Только сборка:

```bash
make build
```

Go-тесты внутренних пакетов с детектором гонок (отдельный сервер и PostgreSQL не нужны):

```bash
make test
```

Если локальный Go build cache недоступен для записи:

```bash
GOCACHE=/tmp/go-build-cache go test -race ./internal/...
```
