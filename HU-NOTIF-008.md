# Evidencia - HU-NOTIF-008: Levantado local end-to-end

## 1. Explicación y abordaje de la Historia de Usuario

Esta última historia de usuario consolida todo el trabajo. El objetivo aquí no es analizar una sola función, sino verificar que todos los engranajes de nuestra arquitectura hexagonal se acoplen perfectamente en un entorno local (End-to-End).

Al analizar la estructura del proyecto y los comandos necesarios, el despliegue E2E se divide en:
*   **Servicios de Soporte (Infraestructura):** El sistema depende críticamente de componentes externos: **PostgreSQL** para la persistencia, **RabbitMQ** para la mensajería asíncrona y **MailHog** para interceptar correos de prueba. Estos se levantan de forma aislada (típicamente desde el repositorio `docker-infra`).
*   **Procesos de Aplicación:** El microservicio en sí no es un solo programa monolítico. Tiene dos puntos de entrada (archivos `main.go` diferentes en la carpeta `cmd/`):
    1.  `notification-api`: Se encarga exclusivamente de escuchar el tráfico web (HTTP) en el puerto 8080.
    2.  `notification-worker`: Se encarga de conectarse a RabbitMQ para escuchar eventos, enviar los correos y ejecutar el "Relay" del Patrón Outbox en segundo plano.
*   **Interconexión:** El levantamiento E2E demuestra que la API puede escribir en la base de datos y que el Worker puede leerla, procesar tareas pesadas y comunicarse con el mundo exterior (MailHog/RabbitMQ) sin bloquear a la API.

---

## 2. Diagrama de Ecosistema Local

Este esquema ilustra cómo todos los contenedores y procesos se comunican entre sí en tu computadora:

```mermaid
graph TD
    User((Usuario / Postman))
    API[💻 Notification API (Go)]
    Worker[⚙️ Notification Worker (Go)]
    BD[(🗄️ PostgreSQL)]
    MQ((🐇 RabbitMQ))
    Mail[📧 MailHog]
    
    User -- "POST /notifications" --> API
    API -- "Guarda datos" --> BD
    MQ -- "Envía eventos" --> Worker
    Worker -- "Actualiza estado" --> BD
    Worker -- "Envía correos SMTP" --> Mail
    Worker -- "Publica eventos (Outbox)" --> MQ
```

---

## 3. Mejora Propuesta

**Automatización del entorno de desarrollo local (Developer Experience)**
Actualmente, para levantar todo el proyecto de manera funcional, el desarrollador tiene que abrir varias terminales: una para levantar Docker, otra para exportar variables de entorno y correr el API, y una tercera para exportar las mismas variables y correr el Worker. 

*Mi propuesta:* Deberíamos crear un archivo `docker-compose.local.yml` dentro de este mismo repositorio (o un script ejecutable `.ps1` / `Makefile`). Este archivo construiría directamente la API y el Worker basándose en la carpeta `deploy/` y los conectaría automáticamente a la red de infraestructura. Esto reduciría la curva de aprendizaje de un nuevo desarrollador: con un solo comando `make up` tendrían el sistema End-to-End funcionando sin configurar variables manuales.

---

## 4. Demostración de Funcionamiento

Para la gran demostración final (End-to-End), sigue estos pasos:
1.  **Infraestructura:** Ve a la carpeta del `docker-infra` y levanta todo: `docker-compose up -d`.
2.  **Arranca la API:** En la carpeta del microservicio, abre una PowerShell, declara las variables (basado en `Comandos.md`) y ejecuta:
    ```powershell
    $env:NOTIFICATION_DB_DSN = "postgres://design_software_app:change-me-app@localhost:15432/design-software-develop?sslmode=disable"
    go run ./cmd/notification-api
    ```
3.  **Arranca el Worker:** Abre *otra* pestaña de PowerShell, declara las variables de nuevo y ejecuta:
    ```powershell
    $env:NOTIFICATION_DB_DSN = "postgres://design_software_app:change-me-app@localhost:15432/design-software-develop?sslmode=disable"
    go run ./cmd/notification-worker
    ```
4.  **Prueba E2E (El video final):** 
    - Muestra que tienes ambas terminales corriendo sin errores.
    - Haz una petición por API (HU-001) y publica un evento manual en RabbitMQ (HU-002).
    - Muestra cómo la terminal del Worker reacciona procesando los mensajes.
    - Abre MailHog (`localhost:8025`) y muestra el correo final recibido. ¡Con esto demuestras que todas las piezas del rompecabezas encajan a la perfección!

🎥 **[PEGAR AQUÍ EL ENLACE AL VIDEO]**
