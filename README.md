# iurl

Консольный HTTP-клиент с красивым форматированием JSON, поддержкой IPv4/IPv6/IPv8 и множеством фич из curl.

---

## Для пользователей

### Быстрый старт

```bash
# Сборка
go build -o iurl ./cmd/iurl

# Или через make
make build
```

### Базовое использование

```bash
# Простой GET-запрос
iurl https://api.example.com

# POST с JSON
iurl -X POST -d '{"name":"John"}' https://api.example.com/users

# POST через --json (автоматически ставит Content-Type)
iurl --json '{"name":"John"}' https://api.example.com/users

# Только заголовки (HEAD)
iurl -I https://api.example.com
```

### Тело запроса

```bash
# Form-urlencoded (автоматически ставит Content-Type)
iurl -d "name=John&age=30" https://api.example.com

# JSON (автоматически ставит Content-Type)
iurl --json '{"name":"John"}' https://api.example.com

# JSON из файла
iurl --json @data.json https://api.example.com

# Данные из stdin
echo "data" | iurl -d @- https://api.example.com

# Загрузка файла через PUT
iurl -T file.txt https://api.example.com/upload

# URL-кодирование данных
iurl --data-urlencode "name=John Doe" https://api.example.com

# Multipart form-data
iurl -F "name=John" -F "file=@photo.jpg" https://api.example.com/upload
```

### Аутентификация

```bash
# Basic auth
iurl -u user:password https://api.example.com

# Bearer token (через заголовок)
iurl -H "Authorization: Bearer your-token" https://api.example.com
```

### Заголовки

```bash
# Один заголовок
iurl -H "Accept: application/json" https://api.example.com

# Несколько заголовков
iurl -H "Accept: application/json" -H "X-Custom: value" https://api.example.com
```

### Запрос сжатого ответа

```bash
iurl --compressed https://api.example.com
```

### Таймауты

```bash
# Таймаут подключения (5 секунд)
iurl --connect-timeout 5 https://api.example.com

# Общий таймаут запроса (30 секунд)
iurl --max-time 30 https://api.example.com
```

### Редиректы

```bash
# Следовать редиректам
iurl -L https://example.com/redirect

# Лимит редиректов
iurl -L --max-redirs 5 https://example.com/redirect
```

### Вывод

```bash
# Сохранить тело в файл
iurl -o response.json https://api.example.com

# Сохранить с именем из URL
iurl -O https://example.com/file.zip

# Вывести заголовки ответа
iurl -i https://api.example.com

# Формат вывода
iurl -w "%{http_code} %{time_total}s\n" https://api.example.com
```

### IP-адреса

```bash
# IPv4
iurl http://192.168.1.1/api

# IPv6
iurl http://[::1]:8080/health
iurl http://[2001:db8::1]/api

# IPv8 (ASN-dot notation)
iurl 64496.192.0.2.1
iurl http://64496.192.0.2.1/path
```

### Прокси

```bash
iurl -x http://proxy:8080 https://api.example.com
iurl -x http://proxy:8080 --proxy-user user:pass https://api.example.com
```

### Повторные попытки

```bash
# Повторить 3 раза при ошибке
iurl --retry 3 https://unstable.example.com

# Задержка между попытками
iurl --retry 3 --retry-delay 2 https://unstable.example.com
```

### Cookies

```bash
# Загрузить и сохранить cookies
iurl -b cookies.txt -c cookies.txt https://api.example.com/login

# Только загрузить
iurl -b cookies.txt https://api.example.com/protected
```

### Переменные и шаблоны

```bash
# Переменные из файла (JSON, YAML, dotenv)
iurl --vars-file vars.json https://{{host}}/api

# Переменные из командной строки
iurl --var env=prod --var version=1.0 https://api.{{env}}.example.com/{{version}}/data
```

### Фильтрация JSON (jq-like)

```bash
# Извлечь поле
iurl https://api.example.com/users --query '.[0].name'

# Фильтрация
iurl https://api.example.com/logs --query '.[] | select(.level=="error")'

# Агрегация
iurl https://api.example.com/orders --query 'map(.total) | add'
```

### Сравнение ответов

```bash
# Первый запрос — сохраняет baseline
iurl --diff baseline https://api.example.com/status

# Второй запрос — сравнивает с baseline
iurl --diff baseline https://api.example.com/status
```

### Мониторинг

```bash
# Проверять каждые 10 секунд
iurl --watch 10s https://api.example.com/status

# Каждые 2 минуты
iurl --watch 2m https://api.example.com/status
```

### Генерация кода

```bash
# Сгенерировать Python-код
iurl --generate-code python https://api.example.com

# Go
iurl --generate-code go https://api.example.com

# JavaScript
iurl --generate-code js https://api.example.com

# curl
iurl --generate-code curl https://api.example.com
```

### Профили запросов

```bash
# Сохранить профиль
iurl --save my-request.json -X POST --json '{"key":"value"}' https://api.example.com

# Загрузить и выполнить
iurl --load my-request.json

# Загрузить и переопределить URL
iurl --load my-request.json https://other-api.com
```

### История

```bash
# Показать историю
iurl --history

# Поиск по тегу
iurl --tag "auth-test" https://api.example.com/login
iurl --search auth-test

# Повторить запрос из истории
iurl --rerun 42
```

### Интерактивный режим

```bash
# Конструктор запросов
iurl --build

# Сырой HTTP-диалог
iurl --raw-shell https://api.example.com
```

### Конфигурация

```bash
# Файл конфигурации
iurl -K config.txt

# Формат файла:
# url = https://api.example.com
# method = POST
# header = Content-Type: application/json
# output = response.json
# user = admin:secret
```

### Обработка ошибок

```bash
# Тихий режим (без вывода)
iurl -s https://api.example.com

# Тихий режим с выводом ошибок
iurl -s -S https://api.example.com

# Ошибка при HTTP 4xx/5xx
iurl --fail https://api.example.com

# Ошибка с телом ответа
iurl --fail-with-body https://api.example.com
```

### Write-out формат

```bash
# Вывести код ответа и время
iurl -w "%{http_code} %{time_total}s\n" https://api.example.com

# Доступные переменные:
# %{http_code} - HTTP статус-код
# %{http_content_type} - Content-Type ответа
# %{time_total} - общее время запроса
# %{size_download} - размер тела
# %{url_effective} - итоговый URL
# %{remote_ip} - IP сервера
# %{remote_port} - порт сервера
```

---

## Для разработчиков

### Структура проекта

```
iurl/
├── cmd/iurl/
│   ├── main.go           # Точка входа, оркестрация
│   └── main_test.go      # Интеграционные тесты
├── internal/
│   ├── cfg/              # Конфигурация, CLI, шаблоны, переменные
│   │   └── cfg.go
│   ├── http/             # HTTP-клиент, построение запросов
│   │   └── http.go
│   ├── fmt/              # Форматирование JSON с подсветкой
│   │   └── fmt.go
│   ├── cookiejar/        # Cookie jar (Netscape формат)
│   │   ├── cookiejar.go
│   │   └── cookiejar_test.go
│   ├── repl/             # Интерактивный режим (tab-completion)
│   │   └── repl.go
│   ├── query/            # jq-подобная фильтрация JSON
│   │   └── query.go
│   ├── codegen/          # Генерация кода (python/go/js/curl)
│   │   └── codegen.go
│   ├── storage/          # История, профили, diff-кеш
│   │   └── storage.go
│   └── interactive/      # Интерактивный конструктор запросов
│       └── interactive.go
├── Makefile
├── build.sh
└── go.mod
```

### Архитектура

**Поток выполнения:**
1. `cfg.Parse()` — разбор CLI → `Config`
2. `request.New()` — построение `http.Request` из `Config`
3. `http.DoWithResult()` — выполнение запроса → `Result`
4. `fmt.PrettyPrintJSON()` / `CopyRaw()` — вывод результата

**Зависимости:**
- `github.com/chzyer/readline` — tab-completion в REPL
- `golang.org/x/net` — HTTP/2 поддержка
- `gopkg.in/yaml.v3` — YAML переменные

### Сборка

```bash
# Локальная сборка
make build

# Сборка для всех платформ
make all

# Тесты
make test

# Линтер
make lint

# Очистка
make clean
```

### Тесты

```bash
# Все тесты
go test ./...

# С verbose выводом
go test ./... -v

# Конкретный пакет
go test ./internal/config/ -v

# Интеграционные тесты (с httptest)
go test ./cmd/iurl/ -v
```

### Добавление нового флага

1. Добавить поле в `Config` в `internal/cfg/cfg.go`
2. Зарегистрировать флаг в `Parse()`
3. Обработать логику в `cmd/iurl/main.go`
4. Добавить описание в `PrintHelp()`
5. Написать тест

### Добавление нового пакета

1. Создать директорию `internal/<name>/`
2. Реализовать логику
3. Написать тесты
4. Импортировать в `cmd/iurl/main.go`

### Формат cookie-файла

Netscape формат (совместимый с curl):
```
# Netscape HTTP Cookie File
.domain.com	TRUE	/	FALSE	1735689600	session	abc123
```

Формат: `domain \t flag \t path \t secure \t expires \t name \t value`

### Формат профиля запроса

JSON:
```json
{
  "url": "https://api.example.com",
  "method": "POST",
  "headers": {
    "Content-Type": "application/json"
  },
  "body": "{\"key\":\"value\"}",
  "insecure": false,
  "compressed": true
}
```

### Формат истории

JSON в `~/.iurl_history`:
```json
{
  "entries": [
    {
      "id": 1,
      "timestamp": "2026-01-01T12:00:00Z",
      "method": "GET",
      "url": "https://api.example.com",
      "headers": {},
      "tags": ["test"],
      "status": 200
    }
  ]
}
```

### Скрипты автодополнения

Для генерации скриптов автодополнения (bash/zsh/fish) используйте:

```bash
iurl --completion bash > /etc/bash_completion.d/iurl
iurl --completion zsh > ~/.zsh/completions/_iurl
iurl --completion fish > ~/.config/fish/completions/iurl.fish
```

### Кросс-платформенность

Проект работает на:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

История REPL хранится в `~/.iurl_history` (нормализация путей через `os.PathSeparator`).

---

## Лицензия

MIT