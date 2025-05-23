package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yago-123/peer-hub/pkg/store"
)

const (
	ServerReadTimeout  = 5 * time.Second
	ServerWriteTimeout = 5 * time.Second
	ServerIdleTimeout  = 10 * time.Second
	MaxHeaderBytes     = 1 << 20
)

type Server struct {
	handlers   *Handler
	httpServer *http.Server
}

func New(s store.Store) *Server {
	return &Server{
		handlers: NewHandler(s),
	}
}

func (s *Server) Start(addr string) error {
	r := gin.Default()

	// todo(): add API versioning
	r.POST("/register", s.handlers.RegisterHandler)
	r.GET("/peer/:peer_id", s.handlers.LookupHandler)

	s.httpServer = &http.Server{
		Addr:           addr,
		Handler:        r,
		ReadTimeout:    ServerReadTimeout,
		WriteTimeout:   ServerWriteTimeout,
		IdleTimeout:    ServerIdleTimeout,
		MaxHeaderBytes: MaxHeaderBytes,
	}

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}
