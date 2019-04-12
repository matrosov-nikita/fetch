package main

import (
	"encoding/json"
	"errors"
	"fetch/internal"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	"log"
	"net/http"
)

// Handler represents http router for handling API requests.
type Handler struct {
	sc *internal.Scheduler
	router *mux.Router
}

// ErrInvalidRequestBody happens when request body can not be decoded from JSON.
var ErrInvalidRequestBody = errors.New("could not decode request body")
// ErrInvalidId happens when given id can not be parsed from string.
var ErrInvalidId = errors.New("could not parse id of task")

// NewHandler creates new handler.
func NewHandler(sc *internal.Scheduler) *Handler {
	return &Handler{
		sc:sc,
	}
}

// Attach attaches new API handlers to given router.
func (h *Handler) Attach(r *mux.Router) {
	r.HandleFunc("/tasks", h.Create).Methods("POST")
	r.HandleFunc("/tasks", h.GetAll).Methods("GET")
	r.HandleFunc("/tasks/{id}", h.GetById).Methods("GET")
	r.HandleFunc("/tasks/{id}", h.Delete).Methods("DELETE")
}

// RequestTask represents input data for creating task.
type RequestTask struct {
	URL string `json:"url" example:"https://google.ru"`
	Method string `json:"method" example:"GET"`
	Headers map[string]string `json:"headers"`
}

// ResponseCreateResult represents output data after creating task.
type ResponseCreateResult struct {
	ID uuid.UUID `json:"id"`
}

func (h Handler) Create(w http.ResponseWriter, r *http.Request) {
	var t RequestTask

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		h.Error(w, ErrInvalidRequestBody)
		return
	}

	task, err := h.sc.Schedule(t.URL, t.Method, t.Headers)
	if err != nil {
		h.Error(w, err)
		return
	}

	bs, err := json.Marshal(ResponseCreateResult{
		ID: task.ID,
	})

	if err != nil {
		h.Error(w,err)
		return
	}

	_, err = w.Write(bs)
	if err != nil {
		log.Printf("could not write response for task id: %v", task.ID)
	}
}

func (h Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	tasks := h.sc.FindAll()

	responseTasks := make([]*ResponseTask, len(tasks))
	for i,t := range tasks {
		responseTasks[i] = NewResponseTask(t)
	}

	bs, err := json.Marshal(responseTasks)
	if err != nil {
		h.Error(w,err)
		return
	}

	_, err = w.Write(bs)
	if err != nil {
		log.Println("could not write response for tasks")
	}
}

func (h Handler) GetById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	id := uuid.FromStringOrNil(mux.Vars(r)["id"])
	if id == uuid.Nil {
		h.Error(w,ErrInvalidId)
		return
	}

	task, err := h.sc.FindById(id)
	if err != nil {
		h.Error(w,err)
		return
	}
	bs, err := json.Marshal(NewResponseTask(task))
	if err != nil {
		h.Error(w,err)
		return
	}

	_, err = w.Write(bs)
	if err != nil {
		log.Printf("could not write response for task with id: %v", id)
	}
}

func (h Handler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	id := uuid.FromStringOrNil(mux.Vars(r)["id"])
	if id == uuid.Nil {
		h.Error(w,ErrInvalidId)
		return
	}

	h.sc.Delete(id)
	bs, err := json.Marshal(ResponseCreateResult{
		ID: id,
	})

	if err != nil {
		h.Error(w,err)
		return
	}

	_, err = w.Write(bs)
	if err != nil {
		log.Printf("could not write response for deleted task with id: %v", id)
	}
}


func (h Handler) Error(w http.ResponseWriter, e error) {
	err := customError{Error: e.Error()}
	switch e {
	case internal.ErrInvalidTaskUrl, ErrInvalidRequestBody:
		err.statusCode = http.StatusBadRequest
	case internal.ErrTaskNotFound:
		err.statusCode = http.StatusNotFound
	case internal.ErrServiceOverloaded:
		err.statusCode = http.StatusServiceUnavailable
	default:
		err.statusCode = http.StatusInternalServerError
		err.Error = "Internal Server Error"
	}

	bs, _ := json.Marshal(e)
	w.WriteHeader(err.statusCode)
	w.Write(bs)
}


type ResponseTask struct {
	ID uuid.UUID `json:"id"`
	URL string `json:"url"`
	Status string `json:"status"`
	StatusCode int `json:"statusCode,omitempty"`
	Headers map[string][]string `json:"headers,omitempty"`
	ContentLength int64 `json:"contentLength,omitempty"`
	ResponseBody string `json:"responseBody,omitempty"`
	Error string `json:"error,omitempty"`
}

func NewResponseTask(t *internal.Task) *ResponseTask {
	rt :=  &ResponseTask{
		ID:             t.ID,
		Status:         t.Status,
		StatusCode:     t.StatusCode,
		URL: 			t.URL.String(),
		Headers:        t.ResponseHeaders,
		ContentLength:   t.ContentLength,
		ResponseBody:   t.ResponseBody,
	}

	if t.Error != nil {
		rt.Error = t.Error.Error()
	}

	return rt
}

type customError struct {
	Error string `json:"error"`
	statusCode int
}