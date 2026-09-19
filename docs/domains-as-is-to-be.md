# Домены CinemaAbyss: As-Is → To-Be

Текстовая карта доменов — дополнение к C4 Container, не замена. На схеме — контейнеры и связи; здесь таблицы и пояснения, чтобы ревьюеру не расшифровывать note на `.puml`.

Источник для [`docs/c4/02-container-to-be.puml`](c4/02-container-to-be.puml).

**Единая точка входа To-Be:** `proxy-service` (`:8000`). Клиенты ходят только в API Gateway. Sync-маршруты: monolith, movies-service, events-service, billing-orchestrator.

---

## 1. As-Is

Зафиксировано в [ADR-001](adr/0001-as-is-monolith.md) и [`01-context-as-is.puml`](c4/01-context-as-is.puml).

- один процесс Go (`src/monolith/main.go`), порт `:8080`;
- одна PostgreSQL (`src/database/init.sql`), схема `cinemaabyss`;
- синхронный REST → SQL в том же запросе;
- BFF нет: web / mobile / TV бьют в один API.

Публичные маршруты: `GET /health`, `GET/POST /api/users`, `/api/movies`, `/api/payments`, `/api/subscriptions`.

Домены **users**, **movies**, **payments**, **subscriptions** пока логические: общий `main.go`, один пул к БД, пакетных границ нет.

В кейсе есть, в коде нет: login / auth, избранное, оценки как API, скидки, вызов рекомендаций, внешние лояльность / маркетплейсы / платёжные шлюзы. Таблицы `views` и `user_ratings` есть в SQL, HTTP-хендлеров нет.

Compose уже содержит `movies-service`, `proxy-service`, Kafka, `events-service`. Это задел To-Be, не текущий монолит.

---

## 2. Таблица доменов As-Is → To-Be

| Домен / bounded context | As-Is | To-Be container | Sync / async | Статус в репо |
| --- | --- | --- | --- | --- |
| **Users** | monolith, таблица `users` (`password_hash` есть), `/api/users` | **monolith** (context users + authn) | HTTP sync | код monolith; `/api/auth/*` — future |
| **Movies / metadata** | monolith, `movies` + `movie_genres`, `/api/movies` | **movies-service** | HTTP sync; Kafka async (`MovieCreated`, `MovieRated`) | код `src/microservices/movies/` |
| **Payments** | monolith, `payments`, `/api/payments` | **monolith** (шаг Saga вызывает orchestrator) | HTTP sync | код monolith |
| **Subscriptions** | monolith, `subscriptions`, `/api/subscriptions` | **monolith** (шаг Saga вызывает orchestrator) | HTTP sync | код monolith |
| **Billing saga** (оформить подписку) | нет отдельного процесса: payment + subscription в одном API | **billing-orchestrator** + `pg-orchestrator` | HTTP sync между шагами; Kafka async после успеха | нет кода (расширение) |
| **Discounts** | в кейсе, в коде нет | **monolith** (задел) | — | нет |
| **Ratings / views** | таблицы `user_ratings`, `views`; API нет | **monolith** (note «future») | — | SQL есть, хендлеров нет |
| **Domain events** | нет в монолите | **events-service** + **Kafka** | HTTP sync ingestion; Kafka async produce | compose есть; код — задание 2 |
| **Notifications** | в кейсе нет | **notifications-service** + `pg-notifications` | Kafka async → worker → Push / Email | нет кода (расширение) |
| **Recommendations** | внешняя система, async в кейсе | **System_Ext** | Kafka async (future consumer) | внешней системы в репо нет |

Почему так: movies уже выделен в шаблоне (первый Strangler). Users / payments / subscriptions остаются в monolith до следующих итераций. Saga и уведомления — целевой контур на схеме, код позже. Рекомендации остаются внешними, как в кейсе.

---

## 3. Микросервисы To-Be

Container-диаграмма = deployable сервисы + БД + брокер + внешние системы. Bounded context ≠ всегда отдельный микросервис: часть контекстов пока внутри monolith.

| # | Container | Bounded context(s) | Задание 1 | В репо сейчас |
| --- | --- | --- | --- | --- |
| 1 | **proxy-service** `:8000` | API Gateway + **authz** + Strangler (`MOVIES_MIGRATION_PERCENT`) | обязательно | compose; код — задание 2 |
| 2 | **monolith** `:8080` | users (**authn**), payments, subscriptions, discounts | обязательно | код `src/monolith/` |
| 3 | **movies-service** `:8081` | movies / metadata | обязательно | код `src/microservices/movies/` |
| 4 | **events-service** `:8082` | domain events ingestion | обязательно | compose; код — задание 2 |
| 5 | **notifications-service** | notifications | расширение | нет |
| 6 | **billing-orchestrator** | cross-domain saga | расширение | нет |
| — | **Kafka** | шина доменных событий | обязательно | compose (`movie-events`, `user-events`, `payment-events`) |
| — | **PostgreSQL** | одна БД `cinemaabyss` (monolith и movies делят схему); далее db-per-service | обязательно | `src/database/init.sql` |
| — | **pg-orchestrator** | state саги | расширение | нет |
| — | **pg-notifications** | inbox, delivery outbox, DLQ metadata | расширение | нет |
| — | **Redis** | rate limit + Pub/Sub realtime | расширение | нет |
| — | **Recommendations** | внешний контекст | обязательно | System_Ext, не наш сервис |

Итого на схеме минимум: proxy, monolith, movies, events + Kafka + PostgreSQL. Расширение — ещё notifications, billing-orchestrator, Redis.

**events-service** = REST ingestion `/api/events/*` (задание 2) + producer в Kafka.  
**notifications-service** = отдельный consumer; на схеме задания 1 рисуем, код не обязателен.

### Sync vs async

| Тип | Когда | Как | Примеры |
| --- | --- | --- | --- |
| **Sync** | пользователь ждёт ответ сейчас | клиент → proxy → сервис → SQL → ответ | логин, каталог, оплата, CRUD users |
| **Async** | побочный эффект, другой домен, рассылка | сервис → Kafka → consumer(s) | `MovieCreated` → push; `SubscriptionActivated` → welcome |

HTTP — запросы. Kafka — реакции. Ручная команда пользователя в брокер не идёт: нужен ответ в том же запросе.

| Контейнер | Sync | Async |
| --- | --- | --- |
| **proxy-service** | маршрутизация ко всем API; authz JWT | — |
| **monolith** | `/api/users`, `/api/payments`, `/api/subscriptions` | publish user / payment events (future / via events API) |
| **movies-service** | `/api/movies` | publish `MovieCreated`, `MovieRated` → Kafka |
| **events-service** | `/api/events/*` | produce в Kafka |
| **billing-orchestrator** | charge → activate (sync к monolith) | `SubscriptionActivated` → Kafka после успеха |
| **notifications-service** | — | consume Kafka; delivery worker → FCM / Email |
| **Recommendations** | — | consume Kafka (future) |

---

## 4. Bounded contexts внутри monolith

На Container не рисуем Component (это задание 2+). Контексты внутри monolith подписываем на контейнере, отдельным box не выносим.

```
users          — /api/users, auth login/register (To-Be)
payments       — /api/payments
subscriptions  — /api/subscriptions
discounts      — задел (кейс, в коде нет)
views, ratings — таблицы есть, API future
```

| Context | API / данные | Что остаётся в процессе |
| --- | --- | --- |
| **Users + Auth** | `/api/users`; `/api/auth/*` (future); таблица `users` | личность, login, выдача JWT / session |
| **Payments** | `/api/payments`; таблица `payments` | списание; Idempotency-Key на POST (закладка) |
| **Subscriptions** | `/api/subscriptions`; таблица `subscriptions` | план, период, auto_renew |
| **Discounts** | в кейсе, таблиц нет | note «задел»; отдельный сервис не рисуем |
| **Views / ratings** | `views`, `user_ratings` | note «future API»; не отдельный сервис |

Данные платежей и подписок пока в monolith. Orchestrator вызывает эти API, сам сущности не хранит.

---

## 5. Auth

В кейсе пользователь логинится. В коде As-Is auth нет: `password_hash` в SQL есть, API login нет. На To-Be закладываем, реализация позже.

Как в «Тёплом доме» (ADR-007): **личность в users**, **проверка на Gateway**.

### Карта auth на Container (всё видно, без отдельного auth-service)

Auth **есть на схеме** — но не отдельным box, а подписями на proxy и monolith + маршрут `/api/auth/*`. Container = deployable; login/JWT sign живут в monolith (users), validate — в proxy.

| Часть auth | Контейнер на схеме | Как показать |
| --- | --- | --- |
| login, register, выдача JWT | **monolith** | note «authn»; `/api/auth/*` на связи proxy → monolith |
| validate JWT, 401, whitelist | **proxy-service** | note «authz» |
| `users`, `password_hash` | monolith → PostgreSQL | как сейчас |
| **auth-service** | **не рисуем** | users ещё в monolith; отдельный deployable только для login — лишний |
| **Strangler next** | note / таблица §7 | `user-service` (users **+** authn вместе) |

| Что | Где на To-Be | Паттерн |
| --- | --- | --- |
| **Аутентификация** (login, token) | **monolith**, context users | JWT, `/api/auth/*` (future) |
| **Авторизация** (проверка token) | **proxy-service** | API Gateway до маршрутизации |
| **Данные пользователя** | monolith → PostgreSQL `users` | — |

Подробнее — [ADR-003](adr/0003-auth-jwt-monolith-proxy.md).

---

## 6. Расширения

Поверх минимума задания 1. На схеме — note «future / расширение». Код и отдельные сервисы — следующие итерации после proxy / events. Sync/async + Demetra — [ADR-002](adr/0002-sync-async-kafka-demetra.md); auth — [ADR-003](adr/0003-auth-jwt-monolith-proxy.md); Strangler — ADR-004.

### Notifications, Saga, Redis

| Расширение | Контейнер | Зачем |
| --- | --- | --- |
| Уведомления | **notifications-service** + `pg-notifications` | реакция на доменные события, не блокирует HTTP |
| Saga | **billing-orchestrator** + `pg-orchestrator` | payment → subscription → notify; state и компенсации в одном месте |
| Redis | **Redis** | rate limit на proxy + Pub/Sub «бам» in-app |
| Outbox / Inbox / DLQ | notes на movies и notifications | не потерять событие, отбить дубли, изолировать яд |
| Push / Email / SMS | **System_Ext** | доставка наружу |

**Redis не дублирует Kafka.** Redis — счётчики и эфемерное «сейчас на экране». Kafka — событие обязано дойти.

| Роль Redis | Отличие от Kafka |
| --- | --- |
| Token bucket rate limiting на proxy | счётчики, не доменные события |
| Pub/Sub `user:{id}` → WebSocket | колокольчик можно потерять; при открытии app подтянет из Postgres |

Kafka остаётся для `MovieCreated`, outbox, DLQ и события после успешной Saga (`SubscriptionActivated` → notifications).

### Saga: оформить подписку

Общей distributed TX нет. Оркестрация — primary (сценарий и компенсации в одном месте). Хореография через Kafka — альтернатива, не primary.

1. User → proxy → `POST` purchase subscription.
2. Orchestrator пишет saga `RUNNING` в `pg-orchestrator`.
3. Sync: charge payment в monolith (`Idempotency-Key`).
4. Sync: activate subscription в monolith.
5. Если шаг 4 ок: async `SubscriptionActivated` → Kafka → notifications; saga `COMPLETED`.
6. Если шаг 4 упал: compensate `RefundPayment`; saga `COMPENSATED`.

Уведомление **не** в sync-цепочке. Proxy ходит в orchestrator на «оформить подписку»; CRUD users / payments / subscriptions по-прежнему через monolith.

Реализация позже: orchestrator → `/api/payments` + `/api/subscriptions` в monolith; при split — те же шаги, другие сервисы.

### Сценарий «вышел новый фильм»

Admin → movies-service (sync POST) → TX `movie` + outbox → Kafka `movie-events` → notifications (inbox) → delivery job → worker → FCM / Email. После N retry → DLQ.

notifications-service не принимает запросы пользователя напрямую. Настройки канала позже через monolith API.

### Топики Kafka (из `docker-compose.yml`)

| Топик | Событие | Кто публикует | Кто потребляет |
| --- | --- | --- | --- |
| `movie-events` | `MovieCreated`, `MovieRated` | movies-service, events API | notifications, рекомендации (future) |
| `user-events` | `UserRegistered` | monolith / events API | notifications (welcome), analytics |
| `payment-events` | `PaymentCompleted` | monolith / events API | notifications (чек), billing |

### Паттерны Demetra (закладка на схему)

Опора: [Demetra — паттерны бэкенда, часть 1](https://platform.demetra.site/blog/microservice-patterns). Не все 12 — только уместные.

| # | Паттерн | Где в «Кинобездне» |
| --- | --- | --- |
| 6 | Event-Driven Messaging | movies / monolith / events → Kafka |
| 7 | Transactional Outbox | movies-service: `INSERT movie` + outbox в одной TX |
| 8 | Transactional Inbox | notifications: `inbox_messages(message_id)` |
| 7 (2-й слой) | Outbox доставки | notifications + `pg-notifications`: job «отправить push» |
| 3 | Retry + backoff + jitter | worker доставки → FCM / Email |
| 10 | Dead Letter Queue | `delivery-dlq` после N retry |
| 5 | Работа с внешним API | notifications → Push / Email / SMS; таймаут + отдельный пул |
| 1 | Circuit Breaker | notifications: FCM отдельно от Email |
| 9 | Idempotency-Key | monolith `POST /api/payments` |
| 4 | Rate limiting | proxy (token bucket по userId / API key) |
| 11 | Pub/Sub Redis | proxy (счётчики) + realtime WebSocket |
| 12 | Saga (оркестрация) | billing-orchestrator + `pg-orchestrator` |
| — | Strangler Fig (курс) | proxy + `MOVIES_MIGRATION_PERCENT` |

Три слоя async в notifications: (1) доменное событие + Outbox в movies; (2) Inbox + идемпотентный job; (3) delivery worker + retry + breaker + DLQ. Отдельный «DLQ-сервис» и второй кластер Kafka на схеме задания 1 не рисуем.

---

## 7. Strangler next

Первый вынос — **movies-service** (уже в репо). Остальные контексты подписываем note «следующие кандидаты», отдельными контейнерами на этапе задания 1 не делаем.

| Bounded context | Сейчас на схеме | Следующий шаг Strangler |
| --- | --- | --- |
| Users + Auth | внутри monolith | `user-service` (+ auth) |
| Movies / metadata | **movies-service** | уже вынесен |
| Payments | внутри monolith | `payment-service` |
| Subscriptions | внутри monolith | `subscription-service` |
| Discounts | note в monolith | с payments или monolith |
| Domain events | **events-service** | — |
| Notifications | **notifications-service** | расширение |
| Billing saga | **billing-orchestrator** | расширение |
| Recommendations | System_Ext | остаётся внешним |

Триггер на вынос: отдельный цикл деплоя или отдельная команда. Пока этого нет — контекст живёт в monolith.

Чего на Container **нет** (и это нормально для задания 1): Component внутри сервисов; отдельный box на каждый bounded context; Zookeeper / Kafka UI / K8s / Istio; отдельный BFF; отдельный auth-service; db-per-service как факт (только note).

---

## 8. Ссылки

| Артефакт | Зачем |
| --- | --- |
| [ADR-001](adr/0001-as-is-monolith.md) | As-Is: один процесс, одна БД, sync REST |
| [`01-context-as-is.puml`](c4/01-context-as-is.puml) | C4 Context As-Is |
| [`02-container-to-be.puml`](c4/02-container-to-be.puml) | C4 Container To-Be |
| [ADR-002](adr/0002-sync-async-kafka-demetra.md) | sync HTTP + async Kafka; Outbox, Inbox, delivery, DLQ |
| [ADR-003](adr/0003-auth-jwt-monolith-proxy.md) | authn в monolith (JWT), authz на proxy |
| [ADR README](adr/README.md) | индекс решений |
