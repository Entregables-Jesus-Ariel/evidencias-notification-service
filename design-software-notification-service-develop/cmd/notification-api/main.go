package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	httpadapter "github.com/code-sena/design-software-notification-service/internal/adapter/in/http"
	"github.com/code-sena/design-software-notification-service/internal/adapter/out/persistence"
	"github.com/code-sena/design-software-notification-service/internal/application/usecase"
	"github.com/code-sena/design-software-notification-service/internal/platform/logging"
	otelplatform "github.com/code-sena/design-software-notification-service/internal/platform/otel"
)

func main() {
	// 1) Creamos el contexto base del servicio y su logger estructurado.
	//    Esto nos permite iniciar la app con trazas y logs consistentes.
	ctx := context.Background()
	logger := logging.New()

	// 2) Leemos la DSN de la base de datos desde variables de entorno.
	//    Si no viene configurada, el servicio no puede conectarse a la BD.
	dsn := os.Getenv("NOTIFICATION_DB_DSN")
	if dsn == "" {
		log.Fatal("NOTIFICATION_DB_DSN is required (e.g. postgres://<user>:<password>@<host>:<port>/<db>?sslmode=disable); no default is provided for secrets")
	}

	// 3) Configuramos OpenTelemetry para observabilidad: trazas y métricas.
	//    ServiceName identifica el servicio en Jaeger/OTLP, y el entorno se toma de variables.
	providers, err := otelplatform.Setup(ctx, otelplatform.Config{
		ServiceName:  "notification-api",
		Environment:  envOr("NOTIFICATION_DEPLOYMENT_ENVIRONMENT", "develop"),
		OTLPEndpoint: envOr("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		Insecure:     envOr("OTEL_EXPORTER_OTLP_INSECURE", "true") == "true",
	})
	if err != nil {
		log.Fatalf("failed to set up OpenTelemetry: %v", err)
	}

	// 4) Al salir del programa, cerramos correctamente los proveedores de OTEL.
	//    Esto asegura que se exporten los traces antes de terminar la app.
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := providers.Shutdown(shutdownCtx); err != nil {
			logger.Error("otel shutdown error", "error", err.Error())
		}
	}()

	// 5) Creamos el pool de conexiones a PostgreSQL con tracing activado.
	//    El pool es el canal reutilizable para consultar y guardar notificaciones.
	pool, err := persistence.NewPoolWithTracing(ctx, dsn, providers.TracerProvider)
	if err != nil {
		log.Fatalf("failed to open notification_db pool: %v", err)
	}
	defer pool.Close()

	// 6) Creamos los instrumentos de métricas para medir latencia, errores y uso del servicio.
	metrics, err := otelplatform.NewMetrics(providers.MeterProvider.Meter("notification-api"))
	if err != nil {
		log.Fatalf("failed to create metrics instruments: %v", err)
	}

	// 7) Composition root: armamos la aplicación enlazando infraestructura + casos de uso + HTTP.
	//    Aquí se conectan repositorios, casos de uso y el manejador HTTP.
	repo := persistence.NewPgNotificationRepository(pool)
	tmplRepo := persistence.NewPgTemplateRepository(pool)
	sendUC := usecase.NewSendNotification(repo, tmplRepo)
	getUC := usecase.NewGetNotification(repo)
	h := httpadapter.NewHandler(sendUC,
		httpadapter.WithGetUseCase(getUC),
		httpadapter.WithMetrics(metrics),
		httpadapter.WithLogger(logger),
		httpadapter.WithReadinessChecks(map[string]httpadapter.Pinger{"database": repo}),
	)

	// 8) El servicio queda escuchando en un puerto HTTP.
	//    Si PORT no está definido, usa 8080 por defecto.
	addr := ":" + envOr("PORT", "8080")
	log.Printf("notification-api listening on %s", addr)
	if err := http.ListenAndServe(addr, h.Routes()); err != nil {
		log.Fatal(err)
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
