# Architecture Decision Records

Канон: Michael Nygard. Шаблон: Context / Decision / Alternatives / Consequences.

| ADR | Решение | Status |
| --- | --- | --- |
| [ADR-001](0001-as-is-monolith.md) | As-Is: один монолит, синхронный REST, одна PostgreSQL | Accepted |
| [ADR-002](0002-sync-async-kafka-demetra.md) | To-Be: sync HTTP + async Kafka; Demetra Outbox, Inbox, delivery, DLQ | Accepted |
| [ADR-003](0003-auth-jwt-monolith-proxy.md) | To-Be: authn в monolith (JWT), authz на proxy | Accepted |
