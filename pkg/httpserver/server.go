package httpserver

import (
	"context"
	"net/http"
	"time"
)

const (
	//defaultAddr = ":8080"
	defaultReadTimeOut    = 10 * time.Second
	defaultWriteTimeOut   = 10 * time.Second
	DefaultMaxHeaderBytes = 1 << 20
)

type Server struct {
	httpServer *http.Server
}

func (s *Server) Run(port string, handler http.Handler) error {
	s.httpServer = &http.Server{
		Addr:           ":" + port,
		Handler:        handler,
		ReadTimeout:    defaultReadTimeOut,
		WriteTimeout:   defaultWriteTimeOut,
		MaxHeaderBytes: DefaultMaxHeaderBytes,
	}

	return s.httpServer.ListenAndServe()
}

func (s *Server) ShutDown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
