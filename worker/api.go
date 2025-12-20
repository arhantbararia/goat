package worker

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type API struct {
	Address string
	Port    int
	Worker  *Worker
	Router  *chi.Mux
}

func (a *API) initRouter() {
	a.Router = chi.NewRouter()
	a.Router.Route("/tasks", func(r chi.Router) {
		r.Post("/", a.StartTaskHandler)
		r.Get("/", a.GetTasksHandler)
		r.Route("/{taskID}", func(r chi.Router) {
			r.Delete("/", a.StopTaskHandler)
		})

	})

}

func (a *API) Start() {
	fmt.Println("Initializing Worker API")
	a.initRouter()

	fmt.Printf("Worker API listening %s:%d", a.Address, a.Port)
	http.ListenAndServe(fmt.Sprintf("%s:%d", a.Address, a.Port), a.Router)

}
