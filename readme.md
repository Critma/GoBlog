# API for GoBlog
API сервер для веб-блога GoBlog
## Использованные технологии
1. GO language
2. PostgresSQL
3. JWT
4. Bcrypt
5. Swagger
6. PQ (sql адаптер для запуска sql запросов)
## Настройка
Поменять кофигурацию приложения можно в .env файле, или будут использоваться значения по умолчанию
## Полноценный запуск в докере
```shell
docker compose up
```
## Получить документацию
1. Сгенерировать swagger документацию
```shell
make docs
```
2. Запустить сервер
3. Перейти на /swagger/ (https:[host]/swagger/)