package common

import (
	"bytes"
	"testing"

	"github.com/QuantumNous/new-api/constant"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func captureSysLogWriters(t *testing.T) (*bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	LogWriterMu.Lock()
	oldStdout := gin.DefaultWriter
	oldStderr := gin.DefaultErrorWriter
	gin.DefaultWriter = stdout
	gin.DefaultErrorWriter = stderr
	LogWriterMu.Unlock()
	t.Cleanup(func() {
		LogWriterMu.Lock()
		defer LogWriterMu.Unlock()
		gin.DefaultWriter = oldStdout
		gin.DefaultErrorWriter = oldStderr
	})
	return stdout, stderr
}

func TestSysLogDisabledSuppressesSysLogAndSysError(t *testing.T) {
	oldEnabled := constant.SysLogEnabled
	constant.SysLogEnabled = false
	t.Cleanup(func() {
		constant.SysLogEnabled = oldEnabled
	})
	stdout, stderr := captureSysLogWriters(t)

	SysLog("hidden sys log")
	SysError("hidden sys error")

	require.Empty(t, stdout.String())
	require.Empty(t, stderr.String())
}

func TestSysLogEnabledWritesSysLogAndSysError(t *testing.T) {
	oldEnabled := constant.SysLogEnabled
	constant.SysLogEnabled = true
	t.Cleanup(func() {
		constant.SysLogEnabled = oldEnabled
	})
	stdout, stderr := captureSysLogWriters(t)

	SysLog("visible sys log")
	SysError("visible sys error")

	require.Contains(t, stdout.String(), "[SYS]")
	require.Contains(t, stdout.String(), "visible sys log")
	require.Contains(t, stderr.String(), "[SYS]")
	require.Contains(t, stderr.String(), "visible sys error")
}
