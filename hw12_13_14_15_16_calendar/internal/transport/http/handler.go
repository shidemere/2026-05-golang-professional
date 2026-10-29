package internalhttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/repository/event"
)

func hello(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("Hello from server!"))
}

func (s *Server) createEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var request CreateRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		s.logger.Warn("can't decode incoming event create request: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	l := fmt.Sprintf("got create event request with parameters: CreateEvent={%v}", request)
	s.logger.Info(l)

	if err := s.validator.Struct(request); err != nil {
		if _, ok := errors.AsType[validator.ValidationErrors](err); ok {
			s.logger.Warn("validation error occurred", "validation error", err)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "validation error: %v", err)
			return
		}
		s.logger.Warn("error in validation", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "validation error: %v", err)
		return
	}

	created := FromCreateRequest(request)
	event, err := s.c.CreateEvent(r.Context(), created)
	if err != nil {
		s.logger.Warn("can't create event: ", err)
		writeServiceError(w, err)
		return
	}
	response := ToResponse(event)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		s.logger.Warn("can't encode event: ", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *Server) updateEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var request UpdateRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Warn("can't decode incoming event update request: ", err)
		return
	}
	defer r.Body.Close()

	l := fmt.Sprintf("got update event request with parameters: UpdateEvent={%v}", request)
	s.logger.Info(l)

	if err := s.validator.Struct(request); err != nil {
		if _, ok := errors.AsType[validator.ValidationErrors](err); ok {
			s.logger.Warn("validation error occurred", "validation error", err)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "validation error: %v", err)
			return
		}
		s.logger.Warn("error in validation", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "validation error: %v", err)
		return
	}
	// Pointer fields distinguish missing values; required alone accepts a pointer
	// to a zero time or UUID, so reject those values before calling the service.
	if request.ScheduledAt.IsZero() ||
		request.FinishedAt.IsZero() ||
		*request.ID == uuid.Nil ||
		*request.UserID == uuid.Nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	updated := FromUpdateRequest(request)
	event, err := s.c.UpdateEvent(r.Context(), updated)
	if err != nil {
		s.logger.Warn("can't update event: ", err)
		writeServiceError(w, err)
		return
	}
	response := ToResponse(event)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		s.logger.Warn("can't encode event: ", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *Server) deleteEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Query().Get("id")
	if id == "" {
		s.logger.Warn("id for deleting not specified")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("need to specify id"))
		return
	}

	parsedID, err := uuid.Parse(id)
	if err != nil || parsedID == uuid.Nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	l := fmt.Sprintf("got delete event request with parameters: id={%v}", id)
	s.logger.Info(l)
	err = s.c.DeleteEvent(r.Context(), id)
	if err != nil {
		s.logger.Warn("can't delete event: ", err)
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) getByDay(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	query := r.URL.Query().Get("date")
	if query == "" {
		s.logger.Warn("day not specified")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("need to specify day"))
		return
	}
	day, err := time.Parse(time.DateOnly, query)
	if err != nil {
		s.logger.Warn("error in parsing date: ", err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "incorrect format of time, need to be YYYY-MM-DD")
		return
	}

	events, err := s.c.ListByDay(r.Context(), day)
	if err != nil {
		s.logger.Warn("error with getting by day: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(events) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	responses := ToResponses(events)
	err = json.NewEncoder(w).Encode(responses)
	if err != nil {
		s.logger.Warn("can't encode event: ", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *Server) getByWeek(w http.ResponseWriter, r *http.Request) {
	s.getByPeriod(w, r, "week", s.c.ListByWeek)
}

func (s *Server) getByMonth(w http.ResponseWriter, r *http.Request) {
	s.getByPeriod(w, r, "month", s.c.ListByMonth)
}

func (s *Server) getByPeriod(
	w http.ResponseWriter, r *http.Request, period string,
	list func(context.Context, time.Time) ([]event.Event, error),
) {
	w.Header().Set("Content-Type", "application/json")
	query := r.URL.Query().Get("date")
	if query == "" {
		s.logger.Warn(period + " not specified")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "need to specify start of %s", period)
		return
	}

	date, err := time.Parse(time.DateOnly, query)
	if err != nil {
		s.logger.Warn("error in parsing date: ", err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "incorrect format of time, need to be YYYY-MM-DD")
		return
	}

	events, err := list(r.Context(), date)
	if err != nil {
		log.Warn("error with getting by "+period+": ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(events) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	responses := ToResponses(events)
	err = json.NewEncoder(w).Encode(responses)
	if err != nil {
		s.logger.Warn("can't encode event: ", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// writeServiceError maps domain errors without exposing internal storage details.
func writeServiceError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, event.ErrEventNotFound):
		status = http.StatusNotFound
	case errors.Is(err, event.ErrDateBusy):
		status = http.StatusConflict
	}
	w.WriteHeader(status)
}
