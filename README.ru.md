# Telegram Quote Bot
[🇺🇸 en translation](README.md)

Это Telegram бот, который создает стикеры на основе пересланных ему сообщений.

![пример](example.png)

## Как это работает?

Для создания стикеров используются Headless-Chrome и Selenium, что позволяет создавать красивые стикеры любой сложности с использованием HTML, CSS (и даже) JavaScript.

## Как запустить?

Для запуска бота необходимо установить Docker и следовать инструкциям ниже.

1. Клонируйте репозиторий
2. Скопируйте шаблон файла окружения
    ```shell
    cp ./docker/.env.template ./docker/.env
    ```
3. Отредактируйте файл `./docker/.env` и добавьте токен вашего бота:
    ```shell
    $ cat ./docker/.env
    TOKEN=INSERT_TOKEN_HERE
    ```
4. Запустите бота
    ```shell
    make start
    ```
   Это запустит бота в фоновом режиме.
5. Готово, теперь вы можете пересылать сообщения вашему боту
6. Чтобы остановить бота, выполните:
    ```shell
    make stop
    ```

## Как кастомизировать?

- Вы можете отредактировать `quote.html` для изменения дизайна стикера
- Вы можете заменить `font.ttf` для изменения шрифта текста на стикере
- Вы можете заменить `avatar-placeholder.png` для изменения заглушки аватарки пользователя, когда ее невозможно получить