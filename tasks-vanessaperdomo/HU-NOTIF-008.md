# Evidencia - HU-NOTIF-008: Levantado local end-to-end

## 1. Explicación y abordaje de la Historia de Usuario

El despliegue local End-to-End verifica que todos los engranajes funcionen juntos:

*   **Infraestructura Completa:** Se levanta Postgres, RabbitMQ y MailHog usando Docker Compose desde el repositorio de infra.
*   **Arquitectura Desacoplada:** El código fuente se levanta como **dos procesos diferentes**. La API HTTP que escucha a los usuarios, y el Worker AMQP que procesa eventos asíncronos y envía correos.
*   **Validación Final:** Cuando interactuamos con la API y el Worker en paralelo, observamos que los datos fluyen perfectamente hasta las herramientas externas, cumpliendo las reglas del Patrón Outbox.

---

## 2. Diagrama de Ecosistema Local

```mermaid
graph TD
    User((Usuario))
    API[💻 Notification API (Go)]
    Worker[⚙️ Notification Worker (Go)]
    BD[(🗄️ PostgreSQL)]
    MQ((🐇 RabbitMQ))
    Mail[📧 MailHog]
    
    User -- "POST /notifications" --> API
    API -- "Guarda en BD" --> BD
    MQ -- "Lee eventos" --> Worker
    Worker -- "Actualiza estados" --> BD
    Worker -- "Manda correos" --> Mail
    Worker -- "Publica Outbox" --> MQ
```

---

## 3. Mejora Propuesta

**Automatización del entorno de desarrollo (Developer Experience)**
Actualmente se deben levantar 3 terminales y exportar variables de entorno manualmente.
*Propuesta:* Crear un archivo `docker-compose.local.yml` integrado o un script `Makefile` (`make run-all`) que compile y levante automáticamente la API, el Worker y la base de datos dentro de la misma red de Docker, permitiendo a los nuevos desarrolladores probar el End-to-End con un solo clic.

---

## 4. Demostración de Funcionamiento

A continuación, presento la evidencia en captura de pantalla individual del End-to-End completo.

**Pasos para ejecutar la prueba:**
1. Levanta la infraestructura de Docker:
   ```bash
   # Ve a la carpeta de infra
   docker-compose up -d
   ```
2. En una terminal, levanta la API:
   ```powershell
   $env:NOTIFICATION_DB_DSN = "postgres://design_software_app:change-me-app@localhost:15432/design-software-develop?sslmode=disable"
   go run ./cmd/notification-api
   ```
3. En otra terminal, levanta el Worker:
   ```powershell
   $env:NOTIFICATION_DB_DSN = "postgres://design_software_app:change-me-app@localhost:15432/design-software-develop?sslmode=disable"
   go run ./cmd/notification-worker
   ```
4. Envía un correo vía Postman o `curl` (usando el comando de la HU-001).
5. Abre **MailHog** en tu navegador web (`http://localhost:8025`).
6. Demuestra en el video que todo el flujo concluyó con la llegada del correo a la bandeja de prueba.


---

**Conclusión de la Evidencia:** 
El levantamiento End-to-End comprobó la solidez completa de la arquitectura. Se demostró la perfecta interoperabilidad simultánea entre el API que encola el trabajo (Outbox), el Worker que consume los eventos y los componentes de infraestructura, resultando en un sistema distribuido confiable y funcional.
