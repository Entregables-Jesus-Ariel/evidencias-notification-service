# Informe General - Diseño de Software: Notification Service

**Autor/a:** Vanessa Perdomo 
**Ficha:** 3145555  
**Proyecto:** Notification Service (Arquitectura Hexagonal en Go)

---

## 1. Introducción y Contexto Arquitectónico

Este informe documenta el análisis y apropiación del microservicio **Notification Service**, un sistema distribuido desarrollado en lenguaje **Go**. El objetivo del microservicio es gestionar el envío de notificaciones (vía Email o In-App) de forma altamente disponible y resiliente.

El proyecto está diseñado bajo los principios de la **Arquitectura Hexagonal (Ports & Adapters)**, lo que permite separar estrictamente la lógica de dominio (el modelo del correo y sus plantillas) de los mecanismos de entrega (la base de datos Postgres, el broker RabbitMQ y el servidor SMTP MailHog).

---

## 2. Resumen de Historias de Usuario Analizadas

A lo largo del proyecto, se desglosó el funcionamiento del sistema en 8 Historias de Usuario (HUs) fundamentales:

1. **HU-NOTIF-001 (API HTTP):** El sistema recibe peticiones síncronas a través de un endpoint `POST /notifications`. Para evitar bloqueos, guarda la notificación y responde inmediatamente con un `202 Accepted` (procesamiento en segundo plano).
2. **HU-NOTIF-002 (Consumo AMQP):** El microservicio es reactivo. Escucha de forma asíncrona eventos de dominio externos (ej. `alert.triggered`) en RabbitMQ y procede a armar y enviar las notificaciones correspondientes sin intervención directa del usuario.
3. **HU-NOTIF-003 (Múltiples Canales):** Mediante el **Patrón Composite**, un orquestador determina si la notificación debe enviarse por `EMAIL` (usando SMTP) o si es `IN_APP` (solo se registra en BD), permitiendo agregar nuevos canales a futuro sin romper el código.
4. **HU-NOTIF-004 (Resiliencia):** El sistema es tolerante a fallos. Emplea el **Patrón Outbox** para garantizar que ningún mensaje se pierda si la red falla (reintentos automáticos) y usa un *Constraint* de **Idempotencia** en PostgreSQL para rechazar eventos duplicados.
5. **HU-NOTIF-005 (Consultas):** Expone un endpoint `GET /notifications/{id}` para revisar el estado de una notificación enviada, manejando correctamente los errores de negocio (`404 Not Found`).
6. **HU-NOTIF-006 (Plantillas Dinámicas):** Utiliza un repositorio de plantillas. Al enviar una clave (ej. `ALERT_TRIGGERED`), el servicio de dominio renderiza el asunto y el cuerpo del mensaje inyectando variables dinámicas antes de mandarlo.
7. **HU-NOTIF-007 (Observabilidad):** Incluye endpoints de monitoreo (`/health` y `/ready`) que hacen pings reales a la Base de Datos y RabbitMQ para asegurar que el nodo está sano. Además, propaga Trazas Inmortales de **OpenTelemetry** a través de RabbitMQ.
8. **HU-NOTIF-008 (Despliegue Local):** El ecosistema se integra End-to-End separando responsabilidades: la base de datos y RabbitMQ en contenedores Docker, y la aplicación dividida en dos procesos distintos (API y Worker).

---

## 3. Diagrama de Arquitectura Global

> **Nota para ti:** Dibuja el diagrama de arquitectura usando herramientas como Draw.io o usa el siguiente "Prompt" en una Inteligencia Artificial que genere imágenes o diagramas (como Gamma, ChatGPT Plus o Whimsical).
> **👇 CÓDIGO PLANTUML PARA EL DIAGRAMA:**
> Copia el siguiente código y pégalo en una IA o directamente en [PlantText](https://planttext.com/) o [PlantUML Web](http://www.plantuml.com/plantuml/) para generar el diagrama. Descarga la imagen y ponla aquí.
> 
> ```plantuml
> @startuml
> skinparam componentStyle rectangle
> skinparam backgroundColor #Fdfdfd
> 
> actor "Cliente HTTP" as User
> 
> node "Notification Service (Go)" {
>     component "API HTTP Handler" as API
>     component "Worker AMQP Consumer" as Worker
> }
> 
> database "PostgreSQL" as BD
> queue "RabbitMQ" as MQ
> node "MailHog (SMTP)" as Mail
> 
> User --> API : " POST /notifications"
> API --> BD : " Patrón Outbox"
> MQ --> Worker : " Consume Eventos"
> Worker --> BD : " Idempotencia y Estado"
> Worker --> Mail : " Envia Correo (Composite)"
> Worker --> MQ : " Publica Evento Outbox"
> @enduml
> ```
> 
> Ejemplo de cómo incluir la imagen exportada en el documento: `![Diagrama de Arquitectura](./arquitectura.png)`

**[PEGAR AQUÍ LA IMAGEN DEL DIAGRAMA DE ARQUITECTURA]**

---

## 4. Resumen de Mejoras Técnicas Propuestas

Durante el análisis del código, se identificaron 3 oportunidades clave de mejora arquitectónica:
1. **Enrutamiento a Dead Letter Queue (DLQ):** Automatizar que los mensajes corruptos rechazados en RabbitMQ se desvíen a una cola "muerta" para inspección humana, evitando la pérdida de información crítica.
2. **Validación de Ownership (Seguridad):** Implementar reglas en el endpoint `GET /notifications/{id}` para asegurar que solo el propietario de la notificación (`recipient_id`) pueda consultarla.
3. **Motor Oficial de Plantillas:** Sustituir el reemplazo básico (`strings.ReplaceAll`) por el paquete `text/template` nativo de Go, para habilitar condicionales y bucles dentro de los correos dinámicos.

---

## 5. Guía de Ejecución para Video General (End-to-End)

El siguiente paso a paso agrupa todas las funcionalidades del microservicio en **un solo flujo demostrativo**. Úsalo como guion para grabar tu video general.

### Paso 1: Levantar la Infraestructura
En una terminal (idealmente en la raíz o en la carpeta `design-software-docker-infra-main`), levanta los contenedores:
```bash
docker-compose up -d
```
*(Muestra en el video que Postgres, RabbitMQ y MailHog están corriendo en Docker).*

### Paso 2: Levantar el API
Abre una nueva terminal en la carpeta del código fuente, exporta la variable de conexión y arranca el API:
```powershell
$env:NOTIFICATION_DB_DSN = "postgres://design_software_app:change-me-app@localhost:15432/design-software-develop?sslmode=disable"
go run ./cmd/notification-api
```

### Paso 3: Levantar el Worker
Abre otra terminal, exporta la misma variable y arranca el Worker:
```powershell
$env:NOTIFICATION_DB_DSN = "postgres://design_software_app:change-me-app@localhost:15432/design-software-develop?sslmode=disable"
go run ./cmd/notification-worker
```

### Paso 4: Comprobar Salud (HU-007)
En una cuarta terminal, demuestra que el API está conectada a la BD y RabbitMQ exitosamente:
```bash
curl -i -X GET http://localhost:8080/ready
```
*(Debe devolver un `200 OK`).*

### Paso 5: Enviar Notificación vía API (HU-001 y HU-006)
Copia y pega este comando para inyectar una notificación simulando una plantilla:
```bash
curl -X POST http://localhost:8080/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "recipient_id": "123e4567-e89b-12d3-a456-426614174000",
    "recipient_email": "aprendiz@sena.edu.co",
    "channel": "EMAIL",
    "subject": "Notificacion desde API",
    "template_code": "ALERT_TRIGGERED",
    "template_vars": {
      "alert_type": "Fallo en Producción",
      "ficha": "3145555"
    }
  }'
```
*(Debe devolver un `202 Accepted`). Copia el ID que te devuelve.*

### Paso 6: Consultar Notificación (HU-005)
Usa el ID copiado para consultarla (cambia `EL-ID-COPIADO`):
```bash
curl -X GET http://localhost:8080/notifications/EL-ID-COPIADO
```

### Paso 7: Comprobar el Correo (MailHog) (HU-003 y HU-008)
Ve a tu navegador y abre `http://localhost:8025`.
Muestra que el correo llegó correctamente y que las variables de la plantilla se reemplazaron en el asunto/cuerpo.

### Paso 8: Probar Idempotencia y Asincronismo (HU-002 y HU-004)
Ve a la interfaz de RabbitMQ en tu navegador `http://localhost:15672` (usuario/clave por defecto del docker o admin/admin).
Entra a la pestaña **Exchanges**, selecciona `scheduling-events` y en **Publish message** pon en Routing key: `scheduling.schedule.published` y este JSON:
```json
{
  "event_id": "99999999-9999-9999-9999-999999999999",
  "event_type": "scheduling.schedule.published",
  "source_service": "scheduling-service",
  "payload": {
    "published_by": "123e4567-e89b-12d3-a456-426614174000",
    "schedule_name": "Horario Mañana"
  }
}
```
Dale click a **Publish**. Luego revisa la terminal del Worker y MailHog para ver que llegó.
**Inmediatamente después**, vuelve a darle click a **Publish**. 
Explica en el video que, gracias a la **Idempotencia**, la Base de Datos bloqueó este segundo mensaje porque tiene el mismo `event_id` y no se envió un correo duplicado.

---

## 5. Evidencias de Ejecución de Comandos y Consultas SQL

A continuación se adjuntan las capturas de pantalla que demuestran la correcta ejecución de los comandos (Docker, Go, cURL) y las consultas SQL a la base de datos (PostgreSQL) verificando el registro de notificaciones y la tabla Outbox.

🖼️ **[PEGAR AQUÍ LAS CAPTURAS DE TERMINAL Y BASE DE DATOS]**

---

## 6. Evidencia en Video (Ejecución General)

🎥 **[PEGAR AQUÍ EL ENLACE AL VIDEO GENERAL DE ESTA EJECUCIÓN]**

**Conclusión Final:**
Como se evidenció en la demostración End-to-End, la combinación de una Arquitectura Hexagonal con el uso del Patrón Outbox, RabbitMQ y PostgreSQL da como resultado un microservicio robusto, altamente escalable y extremadamente resiliente ante caídas de red o duplicidad de información.
