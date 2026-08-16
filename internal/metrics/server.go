package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

type PingFunc func(ctx context.Context) error

func NewServer(addr string, registry *prometheus.Registry, version string, startTime time.Time, masters []string, ping PingFunc, logger *slog.Logger) *Server {
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	mux.HandleFunc("GET /healthz", healthzHandler(version, startTime, masters, ping))

	return &Server{
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
		},
		logger: logger,
	}
}

func (s *Server) Start(ctx context.Context) {
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("metrics server failed", "error", err)
		}
	}()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("metrics server shutdown error", "error", err)
		}
	}()
}

type healthResponse struct {
	Status  string   `json:"status"`
	Version string   `json:"version,omitempty"`
	Uptime  string   `json:"uptime,omitempty"`
	Masters []string `json:"masters,omitempty"`
	Error   string   `json:"error,omitempty"`
}

func healthzHandler(version string, startTime time.Time, masters []string, ping PingFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		w.Header().Set("Content-Type", "application/json")

		if err := ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(healthResponse{
				Status: "error",
				Error:  fmt.Sprintf("docker unreachable: %v", err),
			})
			return
		}

		_ = json.NewEncoder(w).Encode(healthResponse{
			Status:  "ok",
			Version: version,
			Uptime:  time.Since(startTime).Truncate(time.Second).String(),
			Masters: masters,
		})
	}
}
