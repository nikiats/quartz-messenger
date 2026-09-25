# Сценарии взаимодействия

В рамках MVP принято, что чат открыт пользователем исключительно на одном устройстве в один момент времени.
События, связанные с добавлением/удалением из чата, удалением самого чата и другими событиями, не связанными с доставкой сообщений, не передаются в режиме реального времени.

## Подключение с аутентификацией

```mermaid
sequenceDiagram
    autonumber
    participant C as Клиент
    participant G as Шлюз соединений
    participant M as Сервис мессенджера

    C->>G: WebSocket handshake
    G-->>C: соединение открыто
    C->>G: token
    G->>M: POST /sessions/validate
    alt сессия валидна
        M-->>G: 200
        G-->>C: соединение аутентифицировано
    else сессия невалидна
        M-->>G: 401
        G-->>C: закрытие соединения
    end
```

## Отправка сообщения

```mermaid
sequenceDiagram
    autonumber
    participant A as Клиент А
    participant G as Шлюз соединений
    participant M as Сервис мессенджера
    participant P as PostgreSQL
    participant B as Клиент Б

    A->>G: сообщение
    G->>M: POST /messages
    M->>P: сохранение сообщения
    P-->>M: сообщение сохранено
    M-->>G: 201
    G-->>A: подтверждение отправки
    M->>G: POST /deliver
    G->>B: сообщение
    G-->>M: 202
```

## Загрузка истории сообщений

Выполняется при просмотре пользователем сообщений в прошедшем времени с пагинацией вверх.

```mermaid
sequenceDiagram
    autonumber
    participant C as Клиент
    participant M as Сервис мессенджера
    participant P as PostgreSQL

    C->>M: GET /chats/{id}/messages?before=X&limit=Z
    M->>P: чтение сообщений
    P-->>M: сообщения
    M-->>C: сообщения
```

## Загрузка новых сообщений

Выполняется при подключении к шлюзу соединений для подгрузки сообщений, пришедших во время отсутствия соединения.

```mermaid
sequenceDiagram
    autonumber
    participant C as Клиент
    participant G as Шлюз соединений
    participant M as Сервис мессенджера
    participant P as PostgreSQL

    C->>G: подключение и аутентификация
    G-->>C: соединение установлено
    C->>M: GET /chats/{id}/messages?after=Y&limit=Z
    M->>P: чтение сообщений
    P-->>M: сообщения
    M-->>C: сообщения
```
