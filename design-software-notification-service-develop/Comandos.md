```powershell

$env:NOTIFICATION_DB_DSN = "postgres://design_software_app:change-me-app@localhost:15432/design-software-develop?sslmode=disable"

```

> go run ./cmd/notification-api


### Comprobar el backend

En otra PowerShell:

```powershell
curl http://localhost:8080/health
```

api/                          → contratos generados (OpenAPI) — el "idioma" HTTP
cmd/
 ├── notification-api/        → arranca el servidor HTTP
 └── notification-worker/     → arranca el consumer AMQP + outbox relay

internal/
 ├── domain/                  🟢 EL CENTRO — no depende de nada externo
 │   ├── model/                  → Notification, Template, Recipient, Outbox, errores
 │   └── service/                → template_renderer (lógica pura, sin BD ni HTTP)
 │
 ├── application/             🔵 CASOS DE USO — orquesta el dominio
 │   ├── port/in/                → interfaces "qué puede pedir el exterior" (SendNotification, GetNotification, ConsumeDomainEvent)
 │   ├── port/out/               → interfaces "qué necesita el dominio del exterior" (Repository, Notifier, RecipientResolver, TemplateRepository)
 │   └── usecase/                → implementación de esos casos de uso
 │
 └── adapter/                 🟠 EL MUNDO EXTERIOR — implementa los ports
     ├── in/http/                 → handler REST (implementa port/in)
     ├── in/amqp/                 → consumer de RabbitMQ (implementa port/in)
     ├── out/persistence/         → repos Postgres (implementa port/out)
     ├── out/notifier/            → SMTP, IN_APP, composite (implementa port/out)
     ├── out/messaging/           → outbox_relay (publica eventos)
     └── out/client/              → resolver de destinatarios (stub)

 platform/                    ⚪ infraestructura transversal (no es del hexágono)
     ├── logging/
     └── otel/                    → traces, metrics