package api

import (
	"context"
	"net/http"
	"time"

	"example.com/permission-selector/internal/config"
	"example.com/permission-selector/internal/metrics"
	"example.com/permission-selector/internal/org"
	"example.com/permission-selector/internal/selector"
	"example.com/permission-selector/internal/store"
	"example.com/permission-selector/internal/workflow"
)

type Server struct {
	config   config.Config
	store    *store.Store
	org      *org.Service
	selector *selector.Service
	workflow *workflow.Service
	mux      *http.ServeMux
	http     *http.Server
}

func NewServer(cfg config.Config, database *store.Store, organization *org.Service, selection *selector.Service, flows *workflow.Service) *Server {
	server := &Server{config: cfg, store: database, org: organization, selector: selection, workflow: flows, mux: http.NewServeMux()}
	server.routes()
	server.http = &http.Server{Addr: cfg.HTTPAddress, Handler: server.logging(server.mux), ReadHeaderTimeout: 3 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second}
	return server
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /health", s.health)
	s.mux.HandleFunc("GET /api/tree", s.tree)
	s.mux.HandleFunc("GET /api/accounts", s.accounts)
	s.mux.HandleFunc("POST /api/requests", s.createRequest)
	s.mux.HandleFunc("GET /api/requests/", s.request)
	s.mux.HandleFunc("POST /api/requests/", s.requestAction)
	s.mux.HandleFunc("GET /api/metrics", s.metrics)
}

func (s *Server) ListenAndServe() error { return s.http.ListenAndServe() }

func (s *Server) Shutdown(ctx context.Context) error { return s.http.Shutdown(ctx) }

func (s *Server) logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		start := time.Now()
		next.ServeHTTP(writer, request)
		_ = start
	})
}

func (s *Server) health(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]any{"status": "ok", "service": "permission-object-selector"})
}

func (s *Server) metrics(writer http.ResponseWriter, request *http.Request) {
	report, err := metrics.BuildReport(s.store)
	if err != nil {
		writeError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, report)
}
