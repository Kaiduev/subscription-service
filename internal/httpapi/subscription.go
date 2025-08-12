package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"subscription/internal/repository"
	"subscription/internal/subscription"
	"subscription/pkg/date"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SubscriptionRepo interface {
	Create(ctx context.Context, s *subscription.Subscription) (uuid.UUID, error)
	Get(ctx context.Context, id uuid.UUID) (*subscription.Subscription, error)
	Update(ctx context.Context, s *subscription.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, f repository.ListFilter) ([]subscription.Subscription, error)
	Summary(ctx context.Context, f repository.SummaryFilter) (int, error)
}

type SubscriptionHandler struct {
	log  *slog.Logger
	repo SubscriptionRepo
}

func NewSubscriptionHandler(log *slog.Logger, repo SubscriptionRepo) *SubscriptionHandler {
	return &SubscriptionHandler{log: log, repo: repo}
}

// Create godoc
// @Summary Создать подписку
// @Description Создаёт запись о подписке
// @Accept json
// @Produce json
// @Param data body subscription.CreateRequest true "данные подписки"
// @Success 201 {object} subscription.Response
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/ [post]
func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req subscription.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if req.ServiceName == "" {
		writeError(w, http.StatusBadRequest, "service_name is required")
		return
	}
	if req.Price < 0 {
		writeError(w, http.StatusBadRequest, "price must be >= 0")
		return
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user_id (uuid expected)")
		return
	}
	start, err := date.ParseMonth(req.StartDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid start_date (YYYY-MM or MM-YYYY)")
		return
	}

	var endPtr *time.Time
	if req.EndDate != nil && *req.EndDate != "" {
		end, err := date.ParseMonth(*req.EndDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid end_date (YYYY-MM or MM-YYYY)")
			return
		}
		if end.Before(start) {
			writeError(w, http.StatusBadRequest, "end_date must be >= start_date")
			return
		}
		endPtr = &end
	}

	s := &subscription.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      userID,
		StartDate:   start,
		EndDate:     endPtr,
	}
	id, err := h.repo.Create(r.Context(), s)
	if err != nil {
		h.log.Error("create subscription failed", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	idStr := id.String()
	resp := subscription.Response{
		ID:          idStr,
		ServiceName: s.ServiceName,
		Price:       s.Price,
		UserID:      s.UserID.String(),
		StartDate:   date.FormatMonth(s.StartDate),
	}
	if s.EndDate != nil {
		es := date.FormatMonth(*s.EndDate)
		resp.EndDate = &es
	}
	writeJSON(w, http.StatusCreated, resp)
}

// Get godoc
// @Summary Получить подписку
// @Param id path string true "ID подписки (uuid)"
// @Produce json
// @Success 200 {object} subscription.Response
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id (uuid expected)")
		return
	}

	s, err := h.repo.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "subscription not found")
			return
		}
		h.log.Error("get subscription failed", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	resp := subscription.Response{
		ID:          s.ID.String(),
		ServiceName: s.ServiceName,
		Price:       s.Price,
		UserID:      s.UserID.String(),
		StartDate:   date.FormatMonth(s.StartDate),
	}
	if s.EndDate != nil {
		es := date.FormatMonth(*s.EndDate)
		resp.EndDate = &es
	}
	writeJSON(w, http.StatusOK, resp)
}

// Update godoc
// @Summary     Обновить подписку
// @Description Обновляет все поля записи по ID
// @Accept      json
// @Produce     json
// @Param       id   path string                     true "ID (uuid)"
// @Param       data body subscription.CreateRequest true "Данные подписки"
// @Success     200  {object} map[string]string
// @Failure     400  {object} map[string]string
// @Failure     500  {object} map[string]string
// @Router      /subscriptions/{id} [put]
func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id (uuid expected)")
		return
	}

	var req subscription.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if req.ServiceName == "" {
		writeError(w, http.StatusBadRequest, "service_name is required")
		return
	}
	if req.Price < 0 {
		writeError(w, http.StatusBadRequest, "price must be >= 0")
		return
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user_id (uuid expected)")
		return
	}
	start, err := date.ParseMonth(req.StartDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid start_date")
		return
	}
	var endPtr *time.Time
	if req.EndDate != nil && *req.EndDate != "" {
		end, err := date.ParseMonth(*req.EndDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid end_date")
			return
		}
		if end.Before(start) {
			writeError(w, http.StatusBadRequest, "end_date must be >= start_date")
			return
		}
		endPtr = &end
	}

	s := &subscription.Subscription{
		ID:          id,
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      userID,
		StartDate:   start,
		EndDate:     endPtr,
	}

	if err := h.repo.Update(r.Context(), s); err != nil {
		h.log.Error("update subscription failed", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// Delete godoc
// @Summary Удалить подписку
// @Param   id path string true "ID (uuid)"
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router  /subscriptions/{id} [delete]
func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id (uuid expected)")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		h.log.Error("delete subscription failed", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// List godoc
// @Summary Список подписок
// @Param user_id query string false "UUID пользователя"
// @Param service_name query string false "Название сервиса (LIKE)"
// @Param limit query int false "Лимит (<=100)"
// @Param offset query int false "Смещение"
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /subscriptions/ [get]
func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	var userID *uuid.UUID
	if v := q.Get("user_id"); v != "" {
		u, err := uuid.Parse(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid user_id")
			return
		}
		userID = &u
	}
	var serviceName *string
	if v := q.Get("service_name"); v != "" {
		serviceName = &v
	}

	limit := 20
	offset := 0
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	items, err := h.repo.List(r.Context(), repository.ListFilter{
		UserID: userID, ServiceName: serviceName, Limit: limit, Offset: offset,
	})
	if err != nil {
		h.log.Error("list subscriptions failed", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	out := make([]subscription.Response, 0, len(items))
	for _, s := range items {
		resp := subscription.Response{
			ID:          s.ID.String(),
			ServiceName: s.ServiceName,
			Price:       s.Price,
			UserID:      s.UserID.String(),
			StartDate:   date.FormatMonth(s.StartDate),
		}
		if s.EndDate != nil {
			es := date.FormatMonth(*s.EndDate)
			resp.EndDate = &es
		}
		out = append(out, resp)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items":  out,
		"limit":  limit,
		"offset": offset,
	})
}

// Summary godoc
// @Summary Сумма за период
// @Description Считает суммарную стоимость всех подписок за период (включительно по месяцам)
// @Param start query string true "Начало периода (YYYY-MM или MM-YYYY)"
// @Param end query string true "Конец периода (YYYY-MM или MM-YYYY)"
// @Param user_id query string false "UUID пользователя"
// @Param service_name query string false "Название сервиса (LIKE)"
// @Produce json
// @Success 200 {object} map[string]int
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/summary [get]
func (h *SubscriptionHandler) Summary(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	startStr := q.Get("start")
	endStr := q.Get("end")
	if startStr == "" || endStr == "" {
		writeError(w, http.StatusBadRequest, "start and end are required (YYYY-MM or MM-YYYY)")
		return
	}
	start, err := date.ParseMonth(startStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid start")
		return
	}
	end, err := date.ParseMonth(endStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid end")
		return
	}
	if end.Before(start) {
		writeError(w, http.StatusBadRequest, "end must be >= start")
		return
	}

	var userID *uuid.UUID
	if v := q.Get("user_id"); v != "" {
		u, err := uuid.Parse(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid user_id")
			return
		}
		userID = &u
	}
	var serviceName *string
	if v := q.Get("service_name"); v != "" {
		serviceName = &v
	}

	total, err := h.repo.Summary(r.Context(), repository.SummaryFilter{
		Start: start, End: end, UserID: userID, ServiceName: serviceName,
	})
	if err != nil {
		h.log.Error("summary failed", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]int{"total": total})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}
