package http

import (
	"encoding/json"
	"fmt"
	"keeper/kit/kitlog"
	"keeper/kit/sender"
	"log/slog"
	"net/http"
	"time"
)

type Server struct {
	srv *http.Server
	mux *http.ServeMux
}

type ServerOption func(*serverOptions)

type serverOptions struct {
	addr         string
	readTimeout  time.Duration
	writeTimeout time.Duration
}

func NewServer(opts ...ServerOption) *Server {
	so := &serverOptions{
		addr:         ":8080",
		readTimeout:  10 * time.Second,
		writeTimeout: 10 * time.Second,
	}
	for _, opt := range opts {
		opt(so)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", notFoundErrorHandler)

	handler := recoverMiddleware(loggingMiddleware(mux))

	return &Server{
		srv: &http.Server{
			Addr:         so.addr,
			Handler:      handler,
			ReadTimeout:  so.readTimeout,
			WriteTimeout: so.writeTimeout,
		},
		mux: mux,
	}
}

func WithAddr(addr string) ServerOption {
	return func(so *serverOptions) {
		so.addr = addr
	}
}

func WithTimeout(readTimeout, writeTimeout time.Duration) ServerOption {
	return func(so *serverOptions) {
		so.readTimeout = readTimeout
		so.writeTimeout = writeTimeout
	}
}

func (s *Server) Handle(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.mux.HandleFunc(pattern, handler)
}

func (s *Server) RegisterHandler(handler Handler) {
	s.mux.HandleFunc(handler.Endpoint(), handler.Handle)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start),
		)
	})
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic recovered", "error", rec)
				err := errorResponse(w, "internal server error", http.StatusInternalServerError)
				kitlog.Error(err, kitlog.WithDescription("failed to send error response"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func jsonResponse(w http.ResponseWriter, data any, status int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		return fmt.Errorf("failed to send json response: %w", err)
	}
	return nil
}

func errorResponse(w http.ResponseWriter, message string, status int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(sender.ErrStatus(message))
	if err != nil {
		return fmt.Errorf("failed to encode error response: %w", err)
	}
	return nil
}

func notFoundErrorHandler(w http.ResponseWriter, r *http.Request) {
	err := errorResponse(w, fmt.Sprintf("Not found: %s %s", r.Method, r.URL.Path), http.StatusNotFound)
	kitlog.Error(err)
}
