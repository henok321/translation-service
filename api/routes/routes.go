package routes

import (
	"log/slog"
	"net/http"

	"github.com/henok321/translation-service/api/handlers"
	"github.com/henok321/translation-service/api/middleware"
	"gorm.io/gorm"
)

type RouteSetup struct {
	database *gorm.DB
	router   *http.ServeMux
}

func SetupRouter(database *gorm.DB) *http.ServeMux {
	instance := RouteSetup{
		database: database,
		router:   http.NewServeMux(),
	}
	instance.setup()

	return instance.router
}

func (app *RouteSetup) endpointWithMiddleware(handler http.Handler) http.Handler {
	return middleware.RequestLogging(slog.LevelDebug, handler)
}

func (app *RouteSetup) setup() {
	// health
	app.router.Handle("GET /health", app.endpointWithMiddleware(http.HandlerFunc(handlers.HealthCheck)))
}
