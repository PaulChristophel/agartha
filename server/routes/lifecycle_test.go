package routes

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/PaulChristophel/agartha/server/config"
	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/logger"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type blockingListener struct {
	acceptStarted chan struct{}
	closed        chan struct{}
}

func (listener *blockingListener) Accept() (net.Conn, error) {
	select {
	case <-listener.acceptStarted:
	default:
		close(listener.acceptStarted)
	}
	<-listener.closed
	return nil, net.ErrClosed
}

func (listener *blockingListener) Close() error {
	select {
	case <-listener.closed:
	default:
		close(listener.closed)
	}
	return nil
}

func (listener *blockingListener) Addr() net.Addr { return testAddr("test") }

type testAddr string

func (address testAddr) Network() string { return string(address) }
func (address testAddr) String() string  { return string(address) }

func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(securityHeaders())
	engine.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	response := httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))

	require.Equal(t, "nosniff", response.Header().Get("X-Content-Type-Options"))
	require.Equal(t, "DENY", response.Header().Get("X-Frame-Options"))
	require.Equal(t, "no-referrer", response.Header().Get("Referrer-Policy"))
	require.NotEmpty(t, response.Header().Get("Content-Security-Policy"))
	require.NotEmpty(t, response.Header().Get("Strict-Transport-Security"))
	require.NotEmpty(t, response.Header().Get("Permissions-Policy"))
}

func TestReadyIsDistinctFromLiveness(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	AddPingRoutes(engine.Group("/"))
	originalDB := db.DB
	db.DB = nil
	t.Cleanup(func() { db.DB = originalDB })

	liveness := httptest.NewRecorder()
	engine.ServeHTTP(liveness, httptest.NewRequest(http.MethodGet, "/ping", nil))
	require.Equal(t, http.StatusOK, liveness.Code)

	readiness := httptest.NewRecorder()
	engine.ServeHTTP(readiness, httptest.NewRequest(http.MethodGet, "/ready", nil))
	require.Equal(t, http.StatusServiceUnavailable, readiness.Code)
}

func TestNewHTTPServerAppliesConfiguredTimeouts(t *testing.T) {
	options := config.HTTPOptions{
		ReadTimeout:       time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      3 * time.Second,
		IdleTimeout:       4 * time.Second,
	}
	server := newHTTPServer("127.0.0.1:0", http.NotFoundHandler(), options)
	require.Equal(t, options.ReadTimeout, server.ReadTimeout)
	require.Equal(t, options.ReadHeaderTimeout, server.ReadHeaderTimeout)
	require.Equal(t, options.WriteTimeout, server.WriteTimeout)
	require.Equal(t, options.IdleTimeout, server.IdleTimeout)
}

func TestServeHTTPShutsDownWhenContextIsCancelled(t *testing.T) {
	testLogger, err := logger.InitLogger(gin.TestMode)
	require.NoError(t, err)
	originalLog := log
	log = testLogger
	t.Cleanup(func() { log = originalLog })
	ctx, cancel := context.WithCancel(context.Background())
	server := &http.Server{Addr: "127.0.0.1:0", Handler: http.NotFoundHandler()}
	listener := &blockingListener{acceptStarted: make(chan struct{}), closed: make(chan struct{})}
	result := make(chan error, 1)
	go func() {
		result <- serveHTTPListener(ctx, server, listener, config.HTTPOptions{ShutdownTimeout: time.Second})
	}()

	select {
	case <-listener.acceptStarted:
	case <-time.After(time.Second):
		t.Fatal("server did not start accepting connections")
	}
	cancel()

	select {
	case err := <-result:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("server did not shut down")
	}
	select {
	case <-listener.closed:
	default:
		t.Fatal("server shutdown did not close the listener")
	}
}
