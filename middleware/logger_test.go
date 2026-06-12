package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func captureGinLoggerWriter(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	common.LogWriterMu.Lock()
	oldWriter := gin.DefaultWriter
	gin.DefaultWriter = buf
	common.LogWriterMu.Unlock()
	t.Cleanup(func() {
		common.LogWriterMu.Lock()
		defer common.LogWriterMu.Unlock()
		gin.DefaultWriter = oldWriter
	})
	return buf
}

func TestGinLogDisabledSkipsAccessLogMiddleware(t *testing.T) {
	oldEnabled := constant.GinLogEnabled
	constant.GinLogEnabled = false
	t.Cleanup(func() {
		constant.GinLogEnabled = oldEnabled
	})
	buf := captureGinLoggerWriter(t)
	gin.SetMode(gin.TestMode)
	server := gin.New()
	SetUpLogger(server)
	server.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	server.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Empty(t, buf.String())
}

func TestGinLogEnabledWritesAccessLog(t *testing.T) {
	oldEnabled := constant.GinLogEnabled
	constant.GinLogEnabled = true
	t.Cleanup(func() {
		constant.GinLogEnabled = oldEnabled
	})
	buf := captureGinLoggerWriter(t)
	gin.SetMode(gin.TestMode)
	server := gin.New()
	SetUpLogger(server)
	server.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	server.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, buf.String(), "[GIN]")
	require.Contains(t, buf.String(), "/ping")
}
