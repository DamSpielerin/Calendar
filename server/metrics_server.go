package server

import (
	"calendar/storage"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
)

func NewMetricsServer(store storage.EventStore) *EventServer {
	ms := new(EventServer)

	ms.Store = store
	ms.UserStore = &storage.Users

	router := http.NewServeMux()
	router.HandleFunc("/metrics/events", ms.TotalEvents)
	router.HandleFunc("/metrics/users", ms.TotalUsers)
	router.HandleFunc("/metrics/requests_per_s", ms.Empty)
	router.HandleFunc("/metrics/requests_per_s_per_u", ms.Empty)
	router.HandleFunc("/metrics/goroutines", ms.TotalGoroutines)
	router.HandleFunc("/metrics/memory", ms.TotalMemory)
	router.HandleFunc("/metrics/cpu", ms.TotalCPU)

	ms.Handler = router

	return ms
}
func (ms *EventServer) Empty(w http.ResponseWriter, r *http.Request) {

}
func (ms *EventServer) TotalEvents(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("Number of events: " + strconv.Itoa(ms.Store.Count())))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (ms *EventServer) TotalUsers(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("Number of users: " + strconv.Itoa(ms.UserStore.Count())))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (ms *EventServer) TotalGoroutines(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("Number of goroutines: " + strconv.Itoa(runtime.NumGoroutine())))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (ms *EventServer) TotalMemory(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Println(m.Alloc)
	w.Header().Set("content-type", jsonContentType)
	err := json.NewEncoder(w).Encode(m)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (ms *EventServer) TotalCPU(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("Number of logical CPUs usable by the current process: " + strconv.Itoa(runtime.NumCPU()) + "\n" + "number of cgo calls made by the current process: " + strconv.Itoa(int(runtime.NumCgoCall()))))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
