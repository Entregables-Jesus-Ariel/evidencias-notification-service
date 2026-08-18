# 🗃️ Base de datos — Notification

El dominio **Notification** se encarga de gestionar la creación, almacenamiento y envío de notificaciones.

## 1. `notification_template`

Guarda las **plantillas que se utilizan para crear notificaciones**.

| Campo              | ¿Qué representa?                    |
| ------------------ | ----------------------------------- |
| `id`               | Identificador único de la plantilla |
| `code`             | Código único de la plantilla        |
| `name`             | Nombre de la plantilla              |
| `channel`          | Canal: `EMAIL` o `IN_APP`           |
| `subject_template` | Asunto de la notificación           |
| `body_template`    | Contenido de la notificación        |
| `is_active`        | Indica si está activa               |
| `created_at`       | Fecha de creación                   |
| `updated_at`       | Fecha de actualización              |

**Ejemplo:** una plantilla para enviar un correo de recuperación de contraseña.

---

## 2. `sent_notification`

Guarda las **notificaciones que se han generado para un usuario y su estado de envío**.

| Campo             | ¿Qué representa?                     |
| ----------------- | ------------------------------------ |
| `id`              | Identificador de la notificación     |
| `recipient_id`    | Usuario que recibe la notificación   |
| `recipient_email` | Correo del destinatario              |
| `channel`         | `EMAIL` o `IN_APP`                   |
| `subject`         | Asunto enviado                       |
| `body_summary`    | Resumen del contenido                |
| `send_status`     | `PENDING`, `SENT` o `FAILED`         |
| `failure_reason`  | Motivo si el envío falla             |
| `template_id`     | Plantilla utilizada                  |
| `source_service`  | Servicio que originó la notificación |
| `source_event_id` | Evento que originó la notificación   |
| `sent_at`         | Fecha y hora del envío               |
| `created_at`      | Fecha de creación                    |

**Ejemplo:** se genera un correo → queda `PENDING` → se envía → pasa a `SENT`.

---

## 3. `outbox`

Guarda los **eventos que deben ser publicados a otros servicios**.

| Campo          | ¿Qué representa?                       |
| -------------- | -------------------------------------- |
| `id`           | Identificador del registro             |
| `event_id`     | Identificador único del evento         |
| `event_type`   | Tipo de evento                         |
| `payload`      | Información del evento en formato JSON |
| `created_at`   | Fecha de creación                      |
| `published_at` | Fecha en que fue publicado             |

**Ejemplo:** otro servicio genera un evento → se guarda en `outbox` → el sistema lo publica → se registra `published_at`.

---

## 🔗 Relación conceptual

```text
notification_template
        │
        │ plantilla utilizada
        ▼
sent_notification
        │
        │ puede originarse por un evento
        ▼
     outbox
```

### 🧠 Para recordarlo fácilmente

* **`notification_template`** → 📄 **¿Cómo debe ser la notificación?**
* **`sent_notification`** → 📬 **¿Qué notificación se envió y cuál es su estado?**
* **`outbox`** → 📤 **¿Qué evento está pendiente de publicar?**

> **Importante:** las tablas se crean inicialmente **sin llaves foráneas**. Las relaciones/FK se agregan posteriormente en `04_alter`.
