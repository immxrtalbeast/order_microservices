# order_microservices

Проект реализует backend интернет-магазина на Go в виде набора микросервисов.

Что уже есть в коде:
- регистрация и логин пользователей;
- JWT-аутентификация через `api-gateway`;
- CRUD для товаров;
- создание, получение, список и удаление заказов;
- асинхронный сценарий резервирования товара через Kafka + Saga;
- трассировка запросов через Jaeger;
- загрузка изображений товаров в Supabase Storage.

<img width="6114" height="3537" alt="mermaid_20250731_2d1d91" src="https://github.com/user-attachments/assets/965dd8a3-facc-4823-a34d-9d60d5b6b027" />

## Архитектура

### Сервисы

- `api-gateway` (`:8080`) - внешний HTTP API на Gin.
- `auth-service` (`:44044`) - регистрация, логин, выдача JWT.
- `inventory-service` (`:44045`) - товары и остатки.
- `order-service` (`:44046`) - заказы и позиции заказа.
- `saga-service` - координация фонового процесса резервирования через Kafka.
- `analytics-service` - пока только заготовка, логика не реализована.

### Инфраструктура

- PostgreSQL - хранение пользователей, товаров, заказов и saga-состояний.
- Kafka - обмен событиями и командами между сервисами.
- Jaeger - distributed tracing.
- Supabase Storage - хранение картинок товаров.

### Как идет запрос

Обычный пользовательский запрос идет так:

1. Клиент отправляет HTTP-запрос в `api-gateway`.
2. `api-gateway` валидирует JWT и проксирует запрос в нужный gRPC-сервис.
3. Сервис пишет данные в PostgreSQL и при необходимости публикует событие в Kafka.

Сценарий создания заказа:

1. `POST /api/v1/order/create-order`
2. `api-gateway` вызывает `order-service` по gRPC.
3. `order-service` сохраняет заказ со статусом `PENDING`.
4. `order-service` публикует `OrderCreatedEvent` в topic `saga-replies`.
5. `saga-service` получает событие, создает запись саги и отправляет `InventoryReserveItemsCommand` в topic `saga-commands`.
6. `inventory-service` резервирует остатки и отправляет обратно `InventoryReservedEvent` или `InventoryReservedEventFailed`.
7. `order-service` слушает `InventoryReservedEvent` и обновляет итоговую сумму заказа.

Важно: в коде уже есть заготовка под компенсацию и отмену заказа, но основной рабочий сценарий сейчас - создание заказа, резерв товара и пересчет суммы. Автоматического перевода статуса заказа из `PENDING` в другой статус в текущем коде нет.

## Внешнее API

Базовый URL:

```text
http://localhost:8080/api/v1
```

Защищенные маршруты требуют JWT:
- либо в cookie `jwt`, которую выставляет `/login`;
- либо в заголовке `Authorization: Bearer <token>`.

### Auth

#### `POST /api/v1/register`

Создает пользователя.

Пример body:

```json
{
  "email": "user@example.com",
  "password": "secret123"
}
```

Пример ответа:

```json
{
  "message": "user created",
  "userID": "3e50f7ca-52b2-4b56-bf33-8e31a44d1f1c"
}
```

#### `POST /api/v1/login`

Логинит пользователя, создает JWT и кладет его в cookie `jwt`.

Пример body:

```json
{
  "email": "user@example.com",
  "password": "secret123"
}
```

Пример ответа:

```json
{
  "message": "login success"
}
```

### Inventory

Все маршруты ниже защищены JWT.

#### `POST /api/v1/inventory/add-good`

Создает товар. Запрос должен быть в формате `multipart/form-data`.

Поля формы:
- `name` - название товара, обязательно;
- `category` - категория, обязательно;
- `description` - описание, опционально;
- `volume` - объем, обязательно;
- `price` - цена, обязательно;
- `quantity_in_stock` - остаток на складе, опционально;
- `image` - файл изображения, обязательно.

Поддерживаемые типы файлов:
- `image/jpeg`
- `image/jpg`
- `image/png`
- `image/webp`
- `image/svg+xml`

Пример:

```bash
curl -X POST "http://localhost:8080/api/v1/inventory/add-good" \
  -H "Authorization: Bearer <jwt>" \
  -F "name=Tea" \
  -F "category=Drinks" \
  -F "description=Green tea" \
  -F "volume=500" \
  -F "price=199" \
  -F "quantity_in_stock=20" \
  -F "image=@./tea.png"
```

Пример ответа:

```json
{
  "message": "good added successfuly"
}
```

#### `GET /api/v1/inventory/goods`

Возвращает список товаров.

Пример ответа:

```json
{
  "goods": [
    {
      "id": "2abbd7c8-e152-4bd2-8dd6-f407db413ab8",
      "name": "Tea",
      "category": "Drinks",
      "image_link": "https://...",
      "description": "Green tea",
      "price": 199,
      "volume": 500,
      "quantity_in_stock": 20
    }
  ]
}
```

#### `PATCH /api/v1/inventory/update-good`

Обновляет товар.

Пример body:

```json
{
  "id": "2abbd7c8-e152-4bd2-8dd6-f407db413ab8",
  "name": "Tea Premium",
  "category": "Drinks",
  "description": "Updated description",
  "price": 249,
  "image_link": "https://...",
  "quantity_in_stock": 15
}
```

Пример ответа:

```json
{
  "message": "good updated successfully"
}
```

#### `DELETE /api/v1/inventory/:id`

Удаляет товар по UUID.

Пример ответа:

```json
{
  "message": "good deleted successfully"
}
```

### Order

Все маршруты ниже защищены JWT.

#### `POST /api/v1/order/create-order`

Создает заказ для пользователя из JWT.

Пример body:

```json
{
  "items": [
    {
      "product_id": "2abbd7c8-e152-4bd2-8dd6-f407db413ab8",
      "quantity": 2
    },
    {
      "product_id": "bb31c2e2-3a5e-495d-bf3f-c6637e4a6b0e",
      "quantity": 1
    }
  ]
}
```

Пример ответа:

```json
{
  "order_id": "96340a5c-e2c0-4662-a4b0-f5825d5ae1e3",
  "status": "PENDING"
}
```

#### `GET /api/v1/order/order/:id`

Возвращает заказ по UUID.

В ответе приходит объект заказа из `order-service`: сам заказ, его `items`, `total`, `status`, `created_at`, `updated_at`.

#### `GET /api/v1/order/list-orders/:id?limit=10&offset=0`

Возвращает список заказов пользователя.

Важно: в текущем коде path-параметр `:id` не используется. Фильтрация идет по `userID` из JWT.

Пример ответа:

```json
{
  "orders": [
    {
      "id": "96340a5c-e2c0-4662-a4b0-f5825d5ae1e3",
      "user_id": "3e50f7ca-52b2-4b56-bf33-8e31a44d1f1c",
      "status": "PENDING",
      "total": 597
    }
  ]
}
```

#### `DELETE /api/v1/order/:id`

Удаляет заказ по UUID.

Пример ответа:

```json
{
  "message": "good deleted successfully"
}
```

## Какие внутренние запросы идут между сервисами

### HTTP -> gRPC

- `api-gateway -> auth-service`
  - `Register(email, password)`
  - `Login(email, password)`
- `api-gateway -> inventory-service`
  - `AddGood(...)`
  - `ListProducts()`
  - `UpdateGood(...)`
  - `DeleteGood(goodID)`
- `api-gateway -> order-service`
  - `CreateOrder(userID, items)`
  - `Order(orderID)`
  - `ListOrders(userID, limit, offset)`
  - `DeleteOrder(orderID)`

### Kafka topics и события

Topic `saga-replies`:
- `OrderCreatedEvent` - публикует `order-service`;
- `InventoryReservedEvent` - публикует `inventory-service`;
- `InventoryReservedEventFailed` - публикует `inventory-service`.

Topic `saga-commands`:
- `InventoryReserveItemsCommand` - публикует `saga-service`;
- в коде также есть заготовки под `OrderCancel` и `ReleaseInventoryCommand`.

## Данные и хранение

Что хранится по сервисам:

- `auth-service` - пользователи (`email`, `pass_hash`).
- `inventory-service` - товары (`name`, `category`, `description`, `image_link`, `price`, `volume`, `quantity_in_stock`).
- `order-service` - заказы и позиции заказа.
- `saga-service` - состояние выполнения саги.

Изображения товаров хранятся отдельно в Supabase Storage, а в базе лежит публичная ссылка.

## Локальный запуск

Проект удобнее запускать через `Taskfile`, а Kafka/Jaeger поднять через Docker.

### 1. Поднять инфраструктуру

```bash
docker compose up -d zookeeper kafka1 jaeger
```

После старта будут доступны:
- Kafka: `localhost:9092`
- Jaeger UI: `http://localhost:16686`

### 2. Подготовить `.env`

Каждый сервис читает свой `.env` из своей директории:

- `cmd/api-gateway/.env`
- `cmd/auth-service/.env`
- `cmd/inventory-service/.env`
- `cmd/order-service/.env`
- `cmd/saga-service/.env`

Переменные, которые используются в коде:

Для `auth-service`:

```env
DB_PASS=...
APP_SECRET=...
```

Для `order-service`, `inventory-service`, `saga-service`:

```env
DB_PASS=...
KAFKA_ADDRESS=localhost:9092
```

Для `api-gateway`:

```env
APP_SECRET=...
SUPABASE_ANON_KEY=...
SUPABASE_SECRET_KEY=...
SUPABASE_STORAGE_ENDPOINT=...
SUPABASE_BUCKET_NAME=...
```

### 3. Запустить все сервисы

```bash
task up
```

Task поднимет:
- `auth-service`
- `inventory-service`
- `order-service`
- `saga-service`
- `api-gateway`

### 4. Проверить доступность

- HTTP API: `http://localhost:8080`
- Jaeger UI: `http://localhost:16686`

## Технологии

- Go 1.24.5
- Gin
- gRPC
- GORM
- PostgreSQL
- Kafka
- OpenTelemetry + Jaeger
- Supabase Storage

## Что важно знать

- `docker-compose.yaml` в текущем виде поднимает инфраструктуру и часть готовых контейнеров, но для полного локального сценария удобнее использовать `task up`.
- Создание заказа работает асинхронно: заказ сначала создается, потом отдельно резервируются остатки и пересчитывается сумма.
- Для защищенных маршрутов можно использовать либо cookie `jwt`, либо `Authorization: Bearer <token>`.
