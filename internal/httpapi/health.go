package httpapi

import (
	"net/http"
)

type Health struct{}

func NewHealth() *Health { return &Health{} }

func (h *Health) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
