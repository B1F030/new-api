package logger

import (
	"bytes"
	"context"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func captureLoggerWriters(t *testing.T) (*bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	common.LogWriterMu.Lock()
	oldStdout := gin.DefaultWriter
	oldStderr := gin.DefaultErrorWriter
	gin.DefaultWriter = stdout
	gin.DefaultErrorWriter = stderr
	common.LogWriterMu.Unlock()
	t.Cleanup(func() {
		common.LogWriterMu.Lock()
		defer common.LogWriterMu.Unlock()
		gin.DefaultWriter = oldStdout
		gin.DefaultErrorWriter = oldStderr
	})
	return stdout, stderr
}

func TestInfoLogDisabledSuppressesNonErrorLogs(t *testing.T) {
	oldInfoEnabled := constant.InfoLogEnabled
	oldDebugEnabled := common.DebugEnabled
	constant.InfoLogEnabled = false
	common.DebugEnabled = true
	t.Cleanup(func() {
		constant.InfoLogEnabled = oldInfoEnabled
		common.DebugEnabled = oldDebugEnabled
	})
	stdout, stderr := captureLoggerWriters(t)

	LogInfo(context.Background(), "hidden info")
	LogWarn(context.Background(), "hidden warn")
	LogDebug(context.Background(), "hidden debug")
	LogError(context.Background(), "visible error")

	require.Empty(t, stdout.String())
	require.NotContains(t, stderr.String(), "hidden warn")
	require.NotContains(t, stderr.String(), "hidden debug")
	require.Contains(t, stderr.String(), "[ERROR]")
	require.Contains(t, stderr.String(), "visible error")
}

func TestInfoLogEnabledWritesInfoAndWarnLogs(t *testing.T) {
	oldInfoEnabled := constant.InfoLogEnabled
	constant.InfoLogEnabled = true
	t.Cleanup(func() {
		constant.InfoLogEnabled = oldInfoEnabled
	})
	stdout, stderr := captureLoggerWriters(t)

	LogInfo(context.Background(), "visible info")
	LogWarn(context.Background(), "visible warn")

	require.Contains(t, stdout.String(), "[INFO]")
	require.Contains(t, stdout.String(), "visible info")
	require.Contains(t, stderr.String(), "[WARN]")
	require.Contains(t, stderr.String(), "visible warn")
}
