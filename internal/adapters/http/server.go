package httpadapter

import (
	"net/http"

	"example.com/img-resizer/internal/config"
	"example.com/img-resizer/internal/domain/ports"
)

// Dependency Injection of service layer and config values
type Server struct {
	resizerSvc        ports.ResizerService
	serverMultiplexer *http.ServeMux
	config            config.Config
}

func NewServerInstance(resizerSvc ports.ResizerService, cfg config.Config) *Server {
	server := &Server{resizerSvc: resizerSvc, serverMultiplexer: http.NewServeMux(), config: cfg}
	server.serverMultiplexer.HandleFunc("/resize", server.Resize)      // Register the sync resize handler
	server.serverMultiplexer.HandleFunc("/resize-async", server.ResizeAsync) // Register the async resize handler
	server.config = cfg
	return server
}

func (s *Server) Serve(addr string) error {
	return http.ListenAndServe(addr, s.serverMultiplexer)
}
