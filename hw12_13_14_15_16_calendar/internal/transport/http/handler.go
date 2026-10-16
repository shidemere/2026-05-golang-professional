package internalhttp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/repository/event"
)

func hello(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("Hello from server!"))
}

func (s *Server) createEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	var request event.CreateRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		s.logger.Warn("can't decode incoming event create request: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	l := fmt.Sprintf("got create event request with parameters: CreateEvent={%v}", request)
	s.logger.Info(l)
	wrapper := s.c.CreateEvent(context.Background(), request)
	if wrapper.Err != "" {
		s.logger.Warn("can't create event: ", wrapper.Err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(wrapper.Resp)
	if err != nil {
		s.logger.Warn("can't encode event: ", wrapper.Err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *Server) updateEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
}

func (s *Server) deleteEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
}

func (s *Server) getByDay(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
}

func (s *Server) getByWeek(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
}

func (s *Server) getByMonth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
}
