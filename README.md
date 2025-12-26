# Fuel Calculation Backend

Бэкенд-сервис для расчёта энергии сгорания топлива. Реализован на Go с использованием фреймворка Gin.

Основные функции:
- Аутентификация и регистрация пользователей (JWT, хранение отозванных токенов в Redis)
- Управление справочником топлива (CRUD)
- Создание и модерация заявок на расчёт энергии сгорания
- Интеграция с PostgreSQL (данные), MinIO (изображения), Redis (чёрный список токенов)
- REST API с 21 методом, соответствует ТЗ курса «Разработка интернет-приложений», МГТУ им. Н.Э. Баумана, 2025

Структура:
- `cmd/` — точка входа
- `internal/` — доменные модели, DTO, хендлеры, сервисы, репозитории, middleware
- `migrations/` — миграции базы данных

Зависимости: Gin, MinIO, go-redis, JWT.

Запуск:
1. Настроить подключение к БД, MinIO, Redis, секрет JWT
2. Выполнить миграции
3. Запустить `go run cmd/main.go`

Ссылки:
- **Лабораторная работа 1** — [бэкенд](https://github.com/RogeReksuby/RT5-51-Kulygin-Fuel-Combustion-Calculation-Backend/tree/front_pages_design)  
  Дизайн интерфейса в Figma, статические HTML-страницы, подключение MinIO.

- **Лабораторная работа 2** — [бэкенд](https://github.com/RogeReksuby/RT5-51-Kulygin-Fuel-Combustion-Calculation-Backend/tree/backend_database)  
  Моделирование БД (PostgreSQL), ER-диаграмма, подключение к бэкенду, логическое удаление.

- **Лабораторная работа 3** — [бэкенд](https://github.com/RogeReksuby/RT5-51-Kulygin-Fuel-Combustion-Calculation-Backend/tree/backend_API)  
  Реализация веб-сервиса (REST API), CRUD для топлива и заявок, расчёт энергии сгорания.

- **Лабораторная работа 4** — [бэкенд](https://github.com/RogeReksuby/RT5-51-Kulygin-Fuel-Combustion-Calculation-Backend/tree/backend_auth)  
  Авторизация (JWT + Redis), Swagger, подготовка ТЗ и расчёт аппаратных требований.

- **Лабораторная работа 5** — [фронтенд](https://github.com/RogeReksuby/RT5-51-Kulygin-Fuel-Combustion-Calculation-Frontend/tree/front_spa)  
  Базовое SPA на React (без Redux), интеграция с API, mock-данные.

- **Лабораторная работа 6**
- - [PWA](https://github.com/RogeReksuby/RT5-51-Kulygin-Fuel-Combustion-Calculation-Frontend/tree/front_redux_pwa)
- - [Tauri](https://github.com/RogeReksuby/RT5-51-Kulygin-Fuel-Combustion-Calculation-Frontend/tree/front_tauri_changes)
- - [Github Pages](https://github.com/RogeReksuby/RT5-51-Kulygin-Fuel-Combustion-Calculation-Frontend/tree/gh-pages)    
  Redux Toolkit, адаптивность, PWA, развёртывание на GitHub Pages, нативное приложение на Tauri.

- **Лабораторная работа 7** — [фронтенд](https://github.com/RogeReksuby/RT5-51-Kulygin-Fuel-Combustion-Calculation-Frontend/tree/front_redux_auth_swag)  
  Авторизация в React, интерфейс потребителя топлива, кодогенерация Axios, redux-thunk.

- **Лабораторная работа 8**
- - [фронтенд](https://github.com/RogeReksuby/RT5-51-Kulygin-Fuel-Combustion-Calculation-Frontend/tree/temp_moderator_branch)
- - [бэкенд](https://github.com/RogeReksuby/RT5-51-Kulygin-Fuel-Combustion-Calculation-Backend/tree/backend_with_async)
- - [сервис для расчета](https://github.com/RogeReksuby/RT5-51-Kulygin-Fuel-Combustion-Calculation-Service/tree/main)  
  Асинхронный сервис (Django), short polling, интерфейс инженера-энергетика.

  Полная система: [бэкенд (Go)](https://github.com/RogeReksuby/RT5-51-Kulygin-Fuel-Combustion-Calculation-Backend),
  [фронтенд (React)](https://github.com/RogeReksuby/RT5-51-Kulygin-Fuel-Combustion-Calculation-Frontend),
  [асинхронный расчёт (Django)](https://github.com/RogeReksuby/RT5-51-Kulygin-Fuel-Combustion-Calculation-Service).
  [Github Pages](https://rogereksuby.github.io/RT5-51-Kulygin-Fuel-Combustion-Calculation-Frontend/).
