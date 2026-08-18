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

```mermaid
%%{init: {'theme': 'base', 'themeVariables': { 'lineColor': '#6C8EBF', 'textColor': '#333333', 'fontFamily': 'Arial'}}}%%
graph TD
    User((👤 Cliente HTTP))
    
    subgraph Microservicio Notification Service
        API[💻 API HTTP Handler]
        Worker[⚙️ Worker AMQP Consumer]
    end
    
    BD[(🗄️ PostgreSQL)]
    MQ((🐇 RabbitMQ))
    Mail[📧 SMTP MailHog]
    
    User -- "POST /notifications (Síncrono)" --> API
    API -- "Patrón Outbox" --> BD
    MQ -- "Eventos (Asíncrono)" --> Worker
    Worker -- "Consulta Idempotencia" --> BD
    Worker -- "Patrón Composite" --> Mail
    Worker -- "Publica Evento Outbox" --> MQ

    style User fill:#D5E8D4,stroke:#82B366,stroke-width:2px
    style API fill:#E1D5E7,stroke:#9673A6,stroke-width:2px
    style Worker fill:#E1D5E7,stroke:#9673A6,stroke-width:2px
    style BD fill:#DAE8FC,stroke:#6C8EBF,stroke-width:2px
    style MQ fill:#FFE6CC,stroke:#D79B00,stroke-width:2px
    style Mail fill:#F8CECC,stroke:#B85450,stroke-width:2px
```


## 4. Resumen de Mejoras Técnicas Propuestas

Durante el análisis del código, se identificaron 3 oportunidades clave de mejora arquitectónica:
1. **Enrutamiento a Dead Letter Queue (DLQ):** Automatizar que los mensajes corruptos rechazados en RabbitMQ se desvíen a una cola "muerta" para inspección humana, evitando la pérdida de información crítica.
2. **Validación de Ownership (Seguridad):** Implementar reglas en el endpoint `GET /notifications/{id}` para asegurar que solo el propietario de la notificación (`recipient_id`) pueda consultarla.
3. **Motor Oficial de Plantillas:** Sustituir el reemplazo básico (`strings.ReplaceAll`) por el paquete `text/template` nativo de Go, para habilitar condicionales y bucles dentro de los correos dinámicos.

---

## 5. Guía de Ejecución para Video General (End-to-End)

El siguiente paso a paso agrupa todas las funcionalidades del microservicio en **un solo flujo demostrativo**. Úsalo como guion para grabar tu video general.

### Paso 1: Levantar la Infraestructura
En una terminal (idealmente en la carpeta `design-software-docker-infra-main`), levanta todos los contenedores (incluyendo RabbitMQ y OpenTelemetry):
```powershell
docker-compose --profile broker --profile observability up -d
```
*(Muestra en el video que Postgres, RabbitMQ, MailHog y OpenTelemetry están corriendo).*

### Paso 2: Levantar el API
Abre una nueva terminal en la carpeta del código fuente, exporta la variable de conexión y arranca el API usando la ruta absoluta de Go:
```powershell
Stop-Process -Id (Get-NetTCPConnection -LocalPort 8080).OwningProcess -Force -ErrorAction SilentlyContinue; $env:NOTIFICATION_DB_DSN = "postgres://design_software_user:change-me@localhost:15432/design-software-develop?sslmode=disable"; $env:NOTIFICATION_AMQP_URL = "amqp://app:app@localhost:5672/"; & "C:\Program Files\Go\bin\go.exe" run ./cmd/notification-api
```

### Paso 3: Levantar el Worker
Abre otra terminal en la carpeta del código fuente y arranca el Worker:
```powershell
$env:NOTIFICATION_DB_DSN = "postgres://design_software_user:change-me@localhost:15432/design-software-develop?sslmode=disable"; $env:NOTIFICATION_AMQP_URL = "amqp://app:app@localhost:5672/"; $env:NOTIFICATION_SMTP_ADDR = "localhost:1025"; & "C:\Program Files\Go\bin\go.exe" run ./cmd/notification-worker
```

### Paso 4: Comprobar Salud (HU-007)
En una cuarta terminal, demuestra que el API está conectada a la BD y RabbitMQ exitosamente (usando `curl.exe` en Windows):
```powershell
curl.exe -i -X GET http://localhost:8080/ready
```
*(Debe devolver un `200 OK`).*

### Paso 5: Enviar Notificación vía API (HU-001 y HU-006)
Copia y pega este comando para inyectar una notificación simulando una plantilla:
```powershell
curl.exe -X POST http://localhost:8080/notifications -H "Content-Type: application/json" -d "{ \"recipient_id\": \"123e4567-e89b-12d3-a456-426614174000\", \"recipient_email\": \"aprendiz@sena.edu.co\", \"channel\": \"EMAIL\", \"subject\": \"Notificacion desde API\", \"template_code\": \"ALERT_TRIGGERED\", \"template_vars\": { \"alert_type\": \"Fallo en Produccion\", \"ficha\": \"3145555\" } }"
```
*(Debe devolver un `202 Accepted`). Copia el ID que te devuelve.*

### Paso 6: Consultar Notificación (HU-005)
Usa el ID copiado para consultarla (cambia `EL-ID-COPIADO`):
```powershell
curl.exe -X GET http://localhost:8080/notifications/EL-ID-COPIADO
```

### Paso 7: Comprobar el Correo (MailHog) (HU-003 y HU-008)
Ve a tu navegador y abre `http://localhost:18025`.
Muestra que el correo llegó correctamente y que las variables de la plantilla se reemplazaron en el asunto/cuerpo.

### Paso 8: Probar Idempotencia y Asincronismo (HU-002 y HU-004)
Ve a la interfaz de RabbitMQ en tu navegador `http://localhost:15672` (usuario: `app` / clave: `app`).
Entra a la pestaña **Exchanges**, selecciona `scheduling-events` y en **Publish message** pon en Routing key: `scheduling.schedule.published` y este JSON:
```json
{
  "event_id": "99999999-9999-9999-9999-999999999999",
  "event_type": "scheduling.schedule.published",
  "source_service": "scheduling-service",
  "timestamp": "2026-08-18T10:00:00Z",
  "version": "1.0",
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

## 5. Evidencias Individuales por Historia de Usuario (Capturas)

A continuación se presentan las evidencias paso a paso que demuestran el cumplimiento de cada Historia de Usuario (HU).

### HU-NOTIF-008: Despliegue Local (Docker)
**Descripción:** El ecosistema completo se despliega separando la infraestructura (Postgres, RabbitMQ, OpenTelemetry) de los servicios (API y Worker).
**Comandos:**
```powershell
docker-compose --profile broker --profile observability up -d
docker ps
```
**Evidencia:**
🖼️ `[PEGAR AQUÍ CAPTURA DE LA TERMINAL MOSTRANDO LOS CONTENEDORES CORRIENDO]`

### HU-NOTIF-007: Observabilidad (Health/Ready)
**Descripción:** El sistema expone endpoints para verificar su estado de salud y su conexión a los componentes externos (BD y Broker).
**Comando:**
```powershell
curl.exe -i -X GET http://localhost:8080/ready
```
**Evidencia:**
🖼️ `[PEGAR AQUÍ CAPTURA DE LA TERMINAL MOSTRANDO EL STATUS OK]`

### HU-NOTIF-001: API HTTP (Síncrono a Asíncrono)
**Descripción:** Recepción de peticiones HTTP POST, guardando en Base de Datos y respondiendo rápido sin bloquear la aplicación.
**Comando:**
```powershell
curl.exe -i -X POST http://localhost:8080/notifications -H "Content-Type: application/json" -d "{\`"recipient_id\`": \`"123e4567-e89b-12d3-a456-426614174000\`", \`"recipient_email\`": \`"aprendiz@sena.edu.co\`", \`"channel\`": \`"EMAIL\`", \`"subject\`": \`"Notificacion desde API\`", \`"template_code\`": \`"ALERT_TRIGGERED\`", \`"template_vars\`": { \`"alert_type\`": \`"Fallo en Produccion\`", \`"ficha\`": \`"3145555\`" }}"
```
**Evidencia:**
🖼️ `[PEGAR AQUÍ CAPTURA DEL CÓDIGO 202 ACCEPTED Y EL JSON QUE DEVUELVE LA API]`

### HU-NOTIF-005: Consultas (GET por ID)
**Descripción:** Consultar el estado de una notificación usando el ID único devuelto en el paso anterior.
**Comando:** *(reemplaza el ID por el tuyo)*
```powershell
curl.exe -i -X GET http://localhost:8080/notifications/<EL-ID-COPIADO>
```
**Evidencia:**
🖼️ `[PEGAR AQUÍ CAPTURA DE LA TERMINAL MOSTRANDO LOS DATOS DE LA NOTIFICACIÓN]`

### HU-NOTIF-002: Consumo AMQP (Worker Reactivo)
**Descripción:** El Worker escucha eventos de dominio directamente desde el bus de mensajes RabbitMQ y los procesa sin intervención HTTP.
**Acción:** Publicar el JSON del evento `scheduling.schedule.published` directo en la interfaz de RabbitMQ.
**Evidencia:**
🖼️ `[PEGAR AQUÍ CAPTURA DEL FORMULARIO DE RABBITMQ CON EL JSON Y EL MENSAJE EN VERDE "Message published"]`

### HU-NOTIF-006: Plantillas Dinámicas
**Descripción:** El sistema toma un evento genérico y renderiza un asunto/cuerpo personalizado usando el motor de plantillas (inyectando la variable "Horario Mañana").
**Evidencia:**
🖼️ `[PEGAR AQUÍ CAPTURA DE LA BANDEJA DE MAILHOG DONDE SE VEA EL ASUNTO CON EL TEXTO REEMPLAZADO "Tu horario Horario Mañana fue publicado"]`

### HU-NOTIF-003: Múltiples Canales (Composite Notifier)
**Descripción:** El orquestador (Patrón Composite) determina el canal y decide usar el Notificador SMTP para mandar un correo electrónico real.
**Evidencia:**
🖼️ `[PEGAR AQUÍ LA MISMA CAPTURA DE MAILHOG O UNA MOSTRANDO EL DETALLE DEL CORREO RECIBIDO]`

### HU-NOTIF-004: Resiliencia (Idempotencia en Base de Datos)
**Descripción:** Evitar enviar y registrar correos duplicados si el mismo evento de RabbitMQ se procesa dos veces.
**Acción:** Darle click nuevamente al botón "Publish message" en RabbitMQ con el mismo JSON. Luego revisar la BD para ver que solo guardó 1 fila.
**Comando:**
```powershell
docker exec ds-develop-postgres-1 psql -U design_software_user -d design-software-develop -c "SELECT id, send_status, source_event_id FROM notification.sent_notification;"
```
**Evidencia:**
🖼️ `[PEGAR AQUÍ CAPTURA DE LA TERMINAL MOSTRANDO LA TABLA DE POSTGRES CON UNA SOLA FILA PARA ESE EVENT_ID]`

## 6. Evidencia en Video (Ejecución General)

🎥 **[PEGAR AQUÍ EL ENLACE AL VIDEO GENERAL DE ESTA EJECUCIÓN]**

**Conclusión Final:**
Como se evidenció en la demostración End-to-End, la combinación de una Arquitectura Hexagonal con el uso del Patrón Outbox, RabbitMQ y PostgreSQL da como resultado un microservicio robusto, altamente escalable y extremadamente resiliente ante caídas de red o duplicidad de información.
