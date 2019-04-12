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

type Handler struct {
	sc *internal.Scheduler
}

var ErrInvalidRequestBody = errors.New("could not decode request body")
var ErrInvalidId = errors.New("invalid id")


func NewHandler(sc *internal.Scheduler) *mux.Router {
	h := &Handler{
		sc:sc,
	}
	r := mux.NewRouter()
	r.HandleFunc("/tasks", h.Create).Methods("POST")
	r.HandleFunc("/tasks", h.GetAll).Methods("GET")
	r.HandleFunc("/tasks/{id}", h.GetById).Methods("GET")
	r.HandleFunc("/tasks/{id}", h.Delete).Methods("DELETE")
	return r
}

type RequestTask struct {
	URL string `json:"url"`
	Method string `json:"method"`
	Headers map[string]string `json:"headers"`
}

type ResponseTask struct {
	ID uuid.UUID `json:"id"`
	URL string `json:"url"`
	Status string `json:"status"`
	StatusCode int `json:"statusCode,omitempty"`
	Headers map[string][]string `json:"headers,omitempty"`
	ContentLength int64 `json:"contentLength,omitempty"`
	ResponseBody string `json:"responseBody,omitempty"`
}

func NewResponseTask(t *internal.Task) *ResponseTask {
	return &ResponseTask{
		ID:             t.ID,
		Status:         t.Status,
		StatusCode:     t.StatusCode,
		URL: 			t.URL.String(),
		Headers:        t.ResponseHeaders,
		ContentLength:   t.ContentLength,
		ResponseBody:   t.ResponseBody,
	}
}

type ResponseCreateResult struct {
	ID uuid.UUID `json:"id"`
}

func (h Handler) Create(w http.ResponseWriter, r *http.Request) {
	var t RequestTask

	w.Header().Set("Content-Type", "application/json")
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
	w.Header().Set("Content-Type", "application/json")
	tasks := h.sc.FindAll()

	var responseTasks []*ResponseTask
	for _, t := range tasks {
		responseTasks = append(responseTasks, NewResponseTask(t))
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
	w.Header().Set("Content-Type", "application/json")
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

}


func (h Handler) Error(w http.ResponseWriter, e error) {
	switch e {
	case internal.ErrInvalidTaskUrl, ErrInvalidRequestBody:
		http.Error(w, e.Error(), http.StatusBadRequest)
	case internal.ErrServiceOverloaded:
		http.Error(w, e.Error(), http.StatusServiceUnavailable)
	default:
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}