package server

import "net/http"

func (s *Server) RegisterRoutes() http.Handler {

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.HelloWorldHandler)
	mux.HandleFunc("/health", s.healthHandler)

	// ********
	// API ROUTES
	// ********
	apiMux := http.NewServeMux()
	mux.Handle("/api/", http.StripPrefix("/api", apiMux))

	// ** V1 **
	v1Mux := http.NewServeMux()
	apiMux.Handle("/v1", http.StripPrefix("/v1", v1Mux))

	v1Mux.HandleFunc("POST /workflows", s.handlDefineWorkflow)


	


	return mux
}
