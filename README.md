# Evidencias - Reconocimiento 

Este directorio contiene las evidencias de la actividad de práctica individual sobre el microservicio `design-software-notification-service`.

## Objetivo
El objetivo de esta actividad fue clonar, leer, ejecutar y **entender de punta a punta** el funcionamiento del microservicio (desarrollado en Go bajo una Arquitectura Hexagonal), **sin modificar el código fuente original**.

## Contenido

Se analizaron 8 Historias de Usuario (HUs). Para cada una se generó un documento Markdown de evidencia (`.md`) independiente, redactado en primera persona y listo para ser entregado, que incluye:
1. La explicación de la implementación trazada a través de las capas de dominio, aplicación y adaptadores.
2. Un diagrama de secuencia (generado con Mermaid) mostrando visualmente el flujo de datos.
3. Una propuesta técnica de mejora sustentada.
4. Las instrucciones paso a paso para ejecutar la demostración local y grabar el video.

### Historias Analizadas:
- [HU-NOTIF-001.md](./tasks-vanessaperdomo/HU-NOTIF-001.md): Enviar notificación vía API (contract-first, POST `/notifications`).
- [HU-NOTIF-002.md](./tasks-vanessaperdomo/HU-NOTIF-002.md): Consumir evento y entregar notificación vía RabbitMQ (AMQP).
- [HU-NOTIF-003.md](./tasks-vanessaperdomo/HU-NOTIF-003.md): Soporte para múltiples canales (Composite Notifier para `EMAIL` e `IN_APP`).
- [HU-NOTIF-004.md](./tasks-vanessaperdomo/HU-NOTIF-004.md): Estrategias de resiliencia (Idempotencia por base de datos, reintentos del Patrón Outbox y DLQ).
- [HU-NOTIF-005.md](./tasks-vanessaperdomo/HU-NOTIF-005.md): Consulta de notificaciones enviadas (GET `/notifications/{id}`).
- [HU-NOTIF-006.md](./tasks-vanessaperdomo/HU-NOTIF-006.md): Uso del servicio de dominio de plantillas de notificación (`template_renderer`).
- [HU-NOTIF-007.md](./tasks-vanessaperdomo/HU-NOTIF-007.md): Pruebas de disponibilidad y observabilidad (`/health`, `/ready`, y continuidad de trazas OTel).
- [HU-NOTIF-008.md](./tasks-vanessaperdomo/HU-NOTIF-008.md): Demostración del levantamiento del entorno local End-to-End.
