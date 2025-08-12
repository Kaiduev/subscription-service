package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

type HealthHandler interface {
	Health(http.ResponseWriter, *http.Request)
}

type SubHandler interface {
	Create(http.ResponseWriter, *http.Request)
	Get(http.ResponseWriter, *http.Request)
	Update(http.ResponseWriter, *http.Request)
	Delete(http.ResponseWriter, *http.Request)
	List(http.ResponseWriter, *http.Request)
	Summary(http.ResponseWriter, *http.Request)
}

func NewRouter(health HealthHandler, sub SubHandler) http.Handler {
	r := chi.NewRouter()

	r.Get("/healthz", health.Health)

	r.Route("/subscriptions", func(r chi.Router) {
		r.Post("/", sub.Create)
		r.Get("/", sub.List)
		r.Get("/summary", sub.Summary)
		r.Get("/{id}", sub.Get)
		r.Put("/{id}", sub.Update)
		r.Delete("/{id}", sub.Delete)
	})
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	return r
}
