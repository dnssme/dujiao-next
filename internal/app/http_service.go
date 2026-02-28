package app

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// HTTPService HTTP 服务封装
type HTTPService struct {
	name   string
	server *http.Server
}

// NewHTTPService 创建 HTTP 服务
// PCI-DSS 6.5.10 — MaxHeaderBytes 限制请求头大小以缓解 DoS。
func NewHTTPService(addr string, handler http.Handler) *HTTPService {
	return &HTTPService{
		name: "http",
		server: &http.Server{
			Addr:              addr,
			Handler:           handler,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			ReadHeaderTimeout: 10 * time.Second,
			IdleTimeout:       120 * time.Second,
			MaxHeaderBytes:    1 << 20, // 1 MB
		},
	}
}

// Name 服务名称
func (s *HTTPService) Name() string {
	if s == nil || s.name == "" {
		return "http"
	}
	return s.name
}

// Start 启动服务
func (s *HTTPService) Start(ctx context.Context) error {
	if s == nil || s.server == nil {
		return errors.New("http server not initialized")
	}
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Stop 停止服务
func (s *HTTPService) Stop(ctx context.Context) error {
	if s == nil || s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}
