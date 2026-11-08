# URL shortener 
**Rest API cервис сокращения ссылок на GO**

## Стек 
- Go
- Chi 
- PostgresSQL
- pgx 
- golang-migrate 
- Docker Compose 

## Начало 

### 1. Клонируйте git репозиторий 

```bash
git clone git@github.com:vvervald/url_shortener.git
```

### 2. Создайте .env файла на основе .env.example 

```bash 
cp .env.example .env 
```

### 3. Запустите приложение 

```bash 
docker compose up --build 
```

> Приложение по умолчанию доступно по адресу https://localhost:5050 

## API 

> Все API-эндпоинты с префиксом /api защищены HTTP Basic Authentication.
> Имя пользователя и пароль задаются в файле .env через переменные BASIC_AUTH_USER и BASIC_AUTH_PASSWORD и передаются приложению при запуске.


### API Examples:

`POST /api/urls`

Request body:
```
{
    "url":"http://example.com"
}
```
Response - **201 Created**:
```
{
    "original_url": "http://example.com",
    "short_url": "localhost:5050/xYWFOo",
    "code": "xYWFOo"
}
```
> Если исходный URL уже существует, сервис возвращает ранее созданную короткую ссылку.

--- 

`GET /api/urls/xYWFOo`

Response - **200 OK**:
```
{
    "original_url": "http://example.com",
    "short_url": "localhost:5050/xYWFOo",
    "code": "xYWFOo"
}
```
--- 

`PATCH /api/urls/xYWFOo`

Request body:
```
{
    "url": "http://example.com2",
    "refresh_short_url": true
}
``` 
Response - **200 OK**:
```
{
    "original_url": "http://example.com2",
    "short_url": "localhost:5050/3oMQA2",
    "code": "3oMQA2"
}
```

---

`DELETE /api/urls/xYWFOo`

Response - **204 No Content** 


