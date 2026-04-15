# URL Shortener

Сервис сокращения URL-адресов, написанный на **Go** с использованием **SQLite** в качестве хранилища.

## Возможности

- Сокращение длинных URL с возможностью указания собственного алиаса
- Редирект по короткому алиасу
- Удаление сокращённых ссылок
- Валидация входных данных
- Базовая аутентификация для защищённых эндпоинтов
- Структурированное логирование (`slog`) с разными форматами для разных окружений
- Хранение данных в SQLite (без внешних зависимостей)

## Стек технологий

- **Go 1.26+**
- **chi** — лёгкий и быстрый HTTP-роутер
- **SQLite** (modernc.org/sqlite — чистая Go-реализация, без CGO)
- **cleanenv** — работа с конфигурацией (YAML + env)
- **validator** — валидация запросов
- **godotenv** — загрузка переменных окружения из `.env`

## Архитектура проекта

```
├── cmd/url-shortener/        # точка входа (main.go)
├── internal/
│   ├── config/               # конфигурация приложения
│   ├── http-server/
│   │   ├── handlers/         # обработчики HTTP-запросов
│   │   │   └── url/
│   │   │       ├── save/     # создание короткой ссылки
│   │   │       ├── redirect/ # редирект
│   │   │       └── delete/   # удаление ссылки
│   │   └── middleware/       # промежуточное ПО (логирование)
│   ├── lib/
│   │   ├── api/              # утилиты для API-клиентов
│   │   ├── random/           # генерация случайных алиасов
│   │   └── logger/           # утилиты для логирования
│   └── storage/              # интерфейс и реализация хранилища (SQLite)
├── config/                   # конфигурационные файлы
├── tests/                    # интеграционные тесты
└── deployment/               # файлы для развёртывания (systemd)
```

## API

| Метод | Путь | Описание | Аутентификация |
|-------|------|----------|----------------|
| `POST` | `/url` | Создать короткую ссылку | Basic Auth |
| `GET` | `/{alias}` | Редирект на исходный URL | Нет |
| `DELETE` | `/{alias}` | Удалить короткую ссылку | Basic Auth |

### Создание ссылки

**Запрос:**

```json
{
  "url": "https://example.com/very/long/url",
  "alias": "mylink"
}
```

Поле `alias` необязательно. Если не указано — генерируется случайная строка длиной 6 символов.

**Ответ:**

```json
{
  "status": "OK",
  "alias": "mylink"
}
```

### Редирект

`GET /{alias}` — возвращает `302 Found` с заголовком `Location`, указывающим на исходный URL.

### Удаление

`DELETE /{alias}` — удаляет ссылку. Возвращает `204 No Content`.

## Конфигурация

Сервис использует YAML-файл конфигурации и переменные окружения. Путь к файлу задаётся через `CONFIG_PATH`.

### Пример конфигурации (`config/prod.yaml`)

```yaml
env: prod
storage_path: "./storage/storage.db"
http_server:
  address: "0.0.0.0:8082"
  timeout: 4s
  idle_timeout: 60s
  user: "deni"
```

### Переменные окружения

| Переменная | Описание |
|------------|----------|
| `CONFIG_PATH` | **Обязательная.** Путь к YAML-файлу конфигурации |
| `HTTP_SERVER_PASSWORD` | **Обязательная.** Пароль для Basic Auth |

Также можно использовать `.env` файл для загрузки переменных окружения.

### Окружения

| Окружение | Формат логов | Уровень |
|-----------|-------------|---------|
| `local` | Pretty (цветной, человекочитаемый) | Debug |
| `dev` | JSON | Debug |
| `prod` | JSON | Info |

## Установка и запуск

### Сборка

```bash
go build -o url-shortener ./cmd/url-shortener
```

### Запуск

```bash
CONFIG_PATH=./config/prod.yaml HTTP_SERVER_PASSWORD=secret ./url-shortener
```

Или с `.env` файлом:

```bash
# .env
CONFIG_PATH=./config/prod.yaml
HTTP_SERVER_PASSWORD=secret
HTTP_SERVER_USER=deni

./url-shortener
```

## Тесты

### Запуск тестов

```bash
go test ./... -v
```

Для интеграционных тестов сервер должен быть запущен на `localhost:8082`.

## Развёртывание

В директории `deployment/` находится пример unit-файла для **systemd**.

### Установка как systemd-сервис

1. Скопируйте бинарный файл и конфигурацию:

```bash
mkdir -p /root/apps/url-shortener
cp url-shortener /root/apps/url-shortener/
cp config/prod.yaml /root/apps/url-shortener/
```

2. Создайте файл окружения `/root/apps/url-shortener/config.env`:

```
CONFIG_PATH=/root/apps/url-shortener/prod.yaml
HTTP_SERVER_PASSWORD=your_password
HTTP_SERVER_USER=your_user
```

3. Установите unit-файл:

```bash
cp deployment/url-shortener.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable url-shortener
systemctl start url-shortener
```

## Структура базы данных

SQLite база содержит одну таблицу:

```sql
CREATE TABLE url (
    id INTEGER PRIMARY KEY,
    alias TEXT NOT NULL UNIQUE,
    url TEXT NOT NULL
);

CREATE INDEX idx_alias ON url(alias);
```

## Лицензия

MIT
