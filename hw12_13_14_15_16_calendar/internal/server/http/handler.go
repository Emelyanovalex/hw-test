package internalhttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Emelyanovalex/hw12_calendar/internal/storage"
)

type handler struct {
	app    Application
	logger Logger
}

// eventDTO is the JSON representation of an event for HTTP requests and responses.
// StartTime is RFC3339; Duration and NotifyBefore are nanoseconds.
type eventDTO struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	StartTime    string `json:"start_time"`
	Duration     int64  `json:"duration"`
	Description  string `json:"description,omitempty"`
	UserID       string `json:"user_id"`
	NotifyBefore int64  `json:"notify_before,omitempty"`
}

type eventResponse struct {
	Event *eventDTO `json:"event"`
}

type eventsResponse struct {
	Events []eventDTO `json:"events"`
}

type errorResponse struct {
	Error string `json:"error"`
}

const dateLayout = "2006-01-02"

func dtoFromEvent(e storage.Event) eventDTO {
	return eventDTO{
		ID:           e.ID,
		Title:        e.Title,
		StartTime:    e.StartTime.Format(time.RFC3339),
		Duration:     int64(e.Duration),
		Description:  e.Description,
		UserID:       e.UserID,
		NotifyBefore: int64(e.NotifyBefore),
	}
}

func eventFromDTO(dto eventDTO) (storage.Event, error) {
	startTime, err := time.Parse(time.RFC3339, dto.StartTime)
	if err != nil {
		return storage.Event{}, fmt.Errorf("invalid start_time (use RFC3339): %w", err)
	}
	return storage.Event{
		ID:           dto.ID,
		Title:        dto.Title,
		StartTime:    startTime,
		Duration:     time.Duration(dto.Duration),
		Description:  dto.Description,
		UserID:       dto.UserID,
		NotifyBefore: time.Duration(dto.NotifyBefore),
	}, nil
}

func (h *handler) createEvent(w http.ResponseWriter, r *http.Request) {
	var dto eventDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	event, err := eventFromDTO(dto)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.app.CreateEvent(r.Context(), event); err != nil {
		if errors.Is(err, storage.ErrDateBusy) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := eventResponse{Event: ptrOf(dtoFromEvent(event))}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *handler) updateEvent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var dto eventDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	event, err := eventFromDTO(dto)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.app.UpdateEvent(r.Context(), id, event); err != nil {
		if errors.Is(err, storage.ErrEventNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, storage.ErrDateBusy) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	event.ID = id
	resp := eventResponse{Event: ptrOf(dtoFromEvent(event))}
	writeJSON(w, http.StatusOK, resp)
}

func (h *handler) deleteEvent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.app.DeleteEvent(r.Context(), id); err != nil {
		if errors.Is(err, storage.ErrEventNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) listEventsForDay(w http.ResponseWriter, r *http.Request) {
	date, err := parseDateParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	events, err := h.app.ListEventsForDay(r.Context(), date)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, toEventsResponse(events))
}

func (h *handler) listEventsForWeek(w http.ResponseWriter, r *http.Request) {
	date, err := parseDateParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	events, err := h.app.ListEventsForWeek(r.Context(), date)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, toEventsResponse(events))
}

func (h *handler) listEventsForMonth(w http.ResponseWriter, r *http.Request) {
	date, err := parseDateParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	events, err := h.app.ListEventsForMonth(r.Context(), date)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, toEventsResponse(events))
}

func parseDateParam(r *http.Request) (time.Time, error) {
	s := r.URL.Query().Get("date")
	if s == "" {
		return time.Time{}, fmt.Errorf("date query parameter is required")
	}
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format (use YYYY-MM-DD): %w", err)
	}
	return t, nil
}

func toEventsResponse(events []storage.Event) eventsResponse {
	dtos := make([]eventDTO, 0, len(events))
	for _, e := range events {
		dtos = append(dtos, dtoFromEvent(e))
	}
	return eventsResponse{Events: dtos}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, errorResponse{Error: msg})
}

func ptrOf[T any](v T) *T {
	return &v
}
