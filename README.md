# kaiten-mcp

MCP-сервер только для чтения из [API Kaiten](https://developers.kaiten.ru/).
Написан на Go, использует чистую архитектуру и ручное внедрение зависимостей.
Сервер работает только через `stdio` и не открывает TCP-порты.

Первая версия предоставляет ровно 23 инструмента чтения. Все запросы к Kaiten
выполняются методом `GET`; сервер не может создавать, изменять, удалять,
перемещать, архивировать или загружать данные.

## Установка

### Готовый бинарник — Go не нужен

После публикации первого GitHub Release установите актуальный бинарник для
Linux или macOS одной командой:

```sh
curl -fsSL https://raw.githubusercontent.com/OctavianAugust1/kaiten-mcp/main/scripts/install.sh | sh
```

Скрипт определяет `amd64` или `arm64`, скачивает подходящий архив GitHub
Release, сверяет его SHA-256 с файлом контрольных сумм и устанавливает
`kaiten-mcp` в `~/.local/bin`. Этот launcher перед запуском считывает
экспортированные переменные из `~/.bashrc`. При необходимости добавьте этот
каталог в `PATH`.
Для установки конкретной версии задайте `KAITEN_MCP_VERSION=vX.Y.Z`; чтобы
сменить каталог установки, задайте `KAITEN_MCP_INSTALL_DIR=/путь/к/каталогу`.

Эта же команда обновляет уже установленный MCP до последнего GitHub Release:

```sh
curl -fsSL https://raw.githubusercontent.com/OctavianAugust1/kaiten-mcp/main/scripts/install.sh | sh
```

Она заменяет только `~/.local/bin/kaiten-mcp` и внутренний бинарник. `.bashrc`
и конфигурация Codex не изменяются; повторно выполнять `codex mcp add` не нужно.

### Сборка из исходного кода

Установите Go 1.26 или новее, настройте переменные и запустите сервер:

```sh
cp .env.example .env
# Укажите KAITEN_BASE_URL и KAITEN_TOKEN в .env.
go run ./cmd/kaiten-mcp
```

`KAITEN_BASE_URL` — HTTPS-адрес API, например
`https://example.kaiten.ru/api/v1`. Значения переменных окружения процесса
имеют приоритет над значениями в `.env`. Для сборки бинарника:

```sh
go build -o kaiten-mcp ./cmd/kaiten-mcp
./kaiten-mcp
```

Процесс принимает и возвращает JSON-RPC через стандартные ввод и вывод.
Диагностика направляется в stderr, поэтому stdout нельзя засорять сообщениями
оболочки или логами.

## Конфигурация

Скопируйте `.env.example` в `.env` и не добавляйте `.env` в Git.

| Переменная | Обязательна | Назначение |
| --- | --- | --- |
| `KAITEN_BASE_URL` | да | HTTPS-адрес API Kaiten без query-параметров, fragment и учётных данных |
| `KAITEN_TOKEN` | да | Bearer-токен Kaiten |
| `KAITEN_REQUEST_TIMEOUT` | нет | Таймаут одной попытки запроса; duration Go, по умолчанию `30s` |
| `KAITEN_DEFAULT_PAGE_LIMIT` | нет | Размер страницы от `1` до `100`; по умолчанию `50` |
| `LOG_LEVEL` | нет | `debug`, `info`, `warn` или `error`; по умолчанию `info` |

## Подключение MCP-клиента

### Codex CLI

Укажите параметры Kaiten в `~/.bashrc` **до** блока, который завершает
неинтерактивную оболочку (`case $- in ... return`). Например:

```sh
export KAITEN_BASE_URL='https://ваш-домен.kaiten.ru/api/v1'
export KAITEN_TOKEN='новый_токен_kaiten'
export KAITEN_REQUEST_TIMEOUT='30s'
export KAITEN_DEFAULT_PAGE_LIMIT='50'
export LOG_LEVEL='error'
```

После установки зарегистрируйте launcher без передачи токена и других
переменных в конфигурацию Codex:

```sh
codex mcp add kaiten -- ~/.local/bin/kaiten-mcp
```

Если сервер уже был добавлен с `--env`, удалите его и добавьте заново:

```sh
codex mcp remove kaiten
codex mcp add kaiten -- ~/.local/bin/kaiten-mcp
```

Перезапустите сессию Codex после изменения `.bashrc`.

### Claude Desktop

Пример конфигурации Claude Desktop:

```json
{
  "mcpServers": {
    "kaiten": {
      "command": "/абсолютный/путь/к/kaiten-mcp",
      "env": {
        "KAITEN_BASE_URL": "https://example.kaiten.ru/api/v1",
        "KAITEN_TOKEN": "замените-на-токен"
      }
    }
  }
}
```

Используйте один способ конфигурации: переменные `env` клиента либо `.env` в
проекте. Не помещайте токен в систему контроля версий или логи.

## Инструменты

Все инструменты объявляют `readOnlyHint: true`, `destructiveHint: false` и
`idempotentHint: true`. Методы и пути API фиксированы сервером.

| Инструмент | Запрос Kaiten | Параметры |
| --- | --- | --- |
| `kaiten_list_cards` | `GET /cards` | Документированные небета-фильтры карточек; см. ниже |
| `kaiten_get_card` | `GET /cards/{card_id}` | `card_id` |
| `kaiten_list_spaces` | `GET /spaces` | `limit`, `offset` |
| `kaiten_get_space` | `GET /spaces/{space_id}` | `space_id` |
| `kaiten_list_boards` | `GET /spaces/{space_id}/boards` | `space_id` |
| `kaiten_get_board` | `GET /boards/{board_id}` | `board_id` |
| `kaiten_list_columns` | `GET /boards/{board_id}/columns` | `board_id` |
| `kaiten_list_subcolumns` | `GET /columns/{column_id}/subcolumns` | `column_id` |
| `kaiten_list_lanes` | `GET /boards/{board_id}/lanes` | `board_id` |
| `kaiten_list_users` | `GET /users` | `limit`, `offset`, `query`, `type`, `access_type_permissions`, `ids`, `include_inactive`, `exclude_directly_added_members_by_entity_uid` |
| `kaiten_get_current_user` | `GET /users/current` | нет |
| `kaiten_list_space_users` | `GET /spaces/{space_id}/users` | `space_id` |
| `kaiten_get_space_user` | `GET /spaces/{space_id}/users/{user_id}` | `space_id`, `user_id` |
| `kaiten_list_card_comments` | `GET /cards/{card_id}/comments` | `card_id` |
| `kaiten_list_card_members` | `GET /cards/{card_id}/members` | `card_id` |
| `kaiten_list_card_types` | `GET /card-types` | нет |
| `kaiten_get_card_type` | `GET /card-types/{type_id}` | `type_id` |
| `kaiten_list_tags` | `GET /tags` | нет |
| `kaiten_list_documents` | `GET /documents` | `query`, `condition`, `fields`, `version`, `limit`, `offset`, `start_position`, `include_search_preview` |
| `kaiten_get_document` | `GET /documents/{document_uid}` | канонический UUID `document_uid` |
| `kaiten_get_document_schema` | `GET /document-schemas/{id}` | `schema_id` |
| `kaiten_list_document_groups` | `GET /document-groups` | нет |
| `kaiten_get_document_group` | `GET /document-groups/{document_group_uid}` | канонический UUID `document_group_uid` |

`kaiten_list_cards` поддерживает следующие стабильные документированные поля:
`space_id`, `board_id`, `column_id`, `lane_id`, `parent_id`, `type_id`,
`member_id`, `tag_id`, `owner_id`, `state`, `query`, `title`, `number`,
`created_at_from`, `created_at_to`, `updated_at_from`, `updated_at_to`,
`due_date_from`, `due_date_to`, `limit`, `offset`, `version`,
`start_position`, `include_archived` и `include_search_preview`.

### Постраничная выдача

За один вызов инструмент запрашивает только одну ограниченную страницу и не
загружает все страницы автоматически. `limit` ограничен значением `100`.
Для offset-эндпоинтов `pagination.next_offset` возвращается, только когда
сервер получил полную запрошенную страницу. Поиск карточек и документов
поддерживает документированные варианты: offset версии 1 либо курсор
`start_position` версии 2; offset нельзя сочетать с курсором версии 2.
Сервер не добавляет недокументированные параметры пагинации к вложенным
эндпоинтам досок, колонок, подколонок, дорожек, комментариев и участников.

## Docker — локальный образ

Docker — альтернативный способ локального запуска. Образ не публикуется в
GHCR или другом registry: соберите target `runtime` локально. Он запускается
непривилегированным пользователем, использует только stdin/stdout и не
публикует порты:

```sh
docker build --target runtime -t kaiten-mcp .
docker run --rm -i --env-file .env kaiten-mcp
```

Команда `docker compose run --rm -i kaiten-mcp` запускает тот же runtime.
Сервис `app` остаётся оболочкой для разработки:

```sh
docker compose run --rm app
```

## Разработка и проверка

```sh
make fmt
make test
make test-race
make vet
make lint
make vuln
make build
```

Аутентифицированный smoke-тест запускается только явно и выполняет лишь
`GET /spaces`:

```sh
KAITEN_BASE_URL=https://example.kaiten.ru/api/v1 \
KAITEN_TOKEN=замените-на-токен \
go test -tags=integration ./internal/adapter/kaiten -run TestIntegrationListSpaces
```

Описание архитектуры и правил выбора источника API находится в
[docs/development.md](docs/development.md). Поведение эндпоинтов основано на
официальной [документации разработчика Kaiten](https://developers.kaiten.ru/).

## Модель безопасности

- Транспорт блокирует любой HTTP-метод, кроме `GET`, ещё до сетевого вызова.
- Обработчики строят только фиксированные проверенные пути: числовые ID должны
  быть положительными, а идентификаторы документов — каноническими UUID.
- Размер ответов ограничен, таймауты обязательны, временные ошибки повторяются
  ограниченное число раз с учётом отмены контекста.
- Bearer-токен используется только адаптером Kaiten и не записывается в логи.
- MCP resources и prompts не регистрируются, поэтому альтернативного канала
  для изменения данных нет.
