package http

import (
	"encoding/json"
	"fmt"
	"keeper/pkg/logger"
	"keeper/transport"
	"log/slog"
	"net/http"
	"time"
)

type Server struct {
	*http.Server
	mux      *http.ServeMux
	id       string
	ingester *transport.Ingester
}

type ServerOption func(*serverOptions)

type serverOptions struct {
	addr         string
	readTimeout  time.Duration
	writeTimeout time.Duration
}

type ExchangeAdapter func(http.ResponseWriter, *http.Request) transport.Exchange

type request[T any] struct {
	req *http.Request
}

type response struct {
	writer   http.ResponseWriter
	senderID string
}

func NewServer(id string, ingester *transport.Ingester, opts ...ServerOption) *Server {
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
		Server: &http.Server{
			Addr:         so.addr,
			Handler:      handler,
			ReadTimeout:  so.readTimeout,
			WriteTimeout: so.writeTimeout,
		},
		mux:      mux,
		id:       id,
		ingester: ingester,
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

func NewExchangeAdapter[T any](s *Server) ExchangeAdapter {
	return func(w http.ResponseWriter, r *http.Request) transport.Exchange {
		return transport.Exchange{
			Request:  &request[T]{req: r},
			Response: &response{writer: w, senderID: s.id},
		}
	}
}

func (s *Server) Handle(pattern string, adapter ExchangeAdapter) {
	s.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		exchange := adapter(w, r)
		s.ingester.Push(exchange)
	})
}

func (r *request[T]) Route() transport.Route {
	return transport.ParseRoute(r.req.URL.Path)
}

func (r *request[T]) Payload() (any, error) {
	var msg transport.Message[T]
	if err := json.NewDecoder(r.req.Body).Decode(&msg); err != nil {
		return nil, fmt.Errorf("decode request payload: type=%T path=%s method=%s: %w",
			msg, r.req.URL.Path, r.req.Method, err)
	}
	return msg, nil
}

func (r *response) Write(payload any) error {
	r.writer.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(r.writer).Encode(transport.NewMessage(r.senderID, payload))
	if err != nil {
		return fmt.Errorf("failed to encode payload: %w", err)
	}
	return nil
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
				logger.LogIfError(err, logger.WithDescription("failed to send error response"))
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
	err := json.NewEncoder(w).Encode(transport.ErrStatus(message))
	if err != nil {
		return fmt.Errorf("failed to encode error response: %w", err)
	}
	return nil
}

func notFoundErrorHandler(w http.ResponseWriter, r *http.Request) {
	err := errorResponse(w, fmt.Sprintf("Not found: %s %s", r.Method, r.URL.Path), http.StatusNotFound)
	logger.LogIfError(err)
}
