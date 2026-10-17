package internalhttp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/shidemere/2026-05-golang-professional/hw12_13_14_15_calendar/internal/repository/event"
)

func hello(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("Hello from server!"))
}

func (s *Server) createEvent(w http.ResponseWriter, r *http.Request) {
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
	var request event.UpdateRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Warn("can't decode incoming event update request: ", err)
		return
	}
	defer r.Body.Close()

	l := fmt.Sprintf("got update event request with parameters: UpdateEvent={%v}", request)
	s.logger.Info(l)
	wrapper := s.c.UpdateEvent(context.Background(), request)
	if wrapper.Err != "" {
		s.logger.Warn("can't update event: ", wrapper.Err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(wrapper.Resp)
	if err != nil {
		s.logger.Warn("can't encode event: ", wrapper.Err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *Server) deleteEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Query().Get("id")
	if id == "" {
		s.logger.Warn("id for deleting not specified")
		w.Write([]byte("need to specify id"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	l := fmt.Sprintf("got delete event request with parameters: id={%v}", id)
	s.logger.Info(l)
	wrapper := s.c.DeleteEvent(context.Background(), id)
	if wrapper.Err != "" {
		s.logger.Warn("can't delete event: ", wrapper.Err)
		w.WriteHeader(http.StatusInternalServerError)
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
		w.Write([]byte("need to specify day"))
		return
	}
	day, err := time.Parse(time.DateOnly, query)
	if err != nil {
		s.logger.Warn("error in parsing date: ", err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "incorrect format of time, need to be YYYY-MM-DD")
		return
	}

	wrapper := s.c.ListByDay(context.Background(), day)
	if wrapper.Err != "" {
		log.Warn("error with getting by day: ", wrapper.Err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(wrapper.Resp) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	fmt.Fprintf(w, "%v", wrapper.Resp)
}

func (s *Server) getByWeek(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	query := r.URL.Query().Get("date")
	if query == "" {
		s.logger.Warn("week not specified")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "need to specify start of week")
		return
	}

	firstDayOfWeek, err := time.Parse(time.DateOnly, query)
	if err != nil {
		s.logger.Warn("error in parsing date: ", err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "incorrect format of time, need to be YYYY-MM-DD")
		return
	}

	wrapper := s.c.ListByWeek(context.Background(), firstDayOfWeek)
	if wrapper.Err != "" {
		log.Warn("error with getting by week: ", wrapper.Err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(wrapper.Resp) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	fmt.Fprintf(w, "%v", wrapper.Resp)
}

func (s *Server) getByMonth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	query := r.URL.Query().Get("date")
	if query == "" {
		s.logger.Warn("month not specified")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "need to specify start of month")
		return
	}

	firstDayOfWeek, err := time.Parse(time.DateOnly, query)
	if err != nil {
		s.logger.Warn("error in parsing date: ", err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "incorrect format of time, need to be YYYY-MM-DD")
		return
	}

	wrapper := s.c.ListByMonth(context.Background(), firstDayOfWeek)
	if wrapper.Err != "" {
		log.Warn("error with getting by month: ", wrapper.Err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(wrapper.Resp) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	fmt.Fprintf(w, "%v", wrapper.Resp)
}
