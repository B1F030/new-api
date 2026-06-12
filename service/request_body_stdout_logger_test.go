package service

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func withRequestBodyStdoutLogConfig(t *testing.T, enabled bool, samplePercent int, maxLineChars int) {
	t.Helper()
	oldEnabled := constant.RequestBodyLogEnabled
	oldSamplePercent := constant.RequestBodyLogSamplePercent
	oldMaxLineChars := constant.RequestBodyLogMaxLineChars
	constant.RequestBodyLogEnabled = enabled
	constant.RequestBodyLogSamplePercent = samplePercent
	constant.RequestBodyLogMaxLineChars = maxLineChars
	t.Cleanup(func() {
		constant.RequestBodyLogEnabled = oldEnabled
		constant.RequestBodyLogSamplePercent = oldSamplePercent
		constant.RequestBodyLogMaxLineChars = oldMaxLineChars
	})
}

func captureRequestBodyStdoutLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	requestBodyStdoutLogMu.Lock()
	oldWriter := requestBodyStdoutLogWriter
	requestBodyStdoutLogWriter = buf
	requestBodyStdoutLogMu.Unlock()
	t.Cleanup(func() {
		requestBodyStdoutLogMu.Lock()
		defer requestBodyStdoutLogMu.Unlock()
		requestBodyStdoutLogWriter = oldWriter
	})
	return buf
}

func newRequestBodyStdoutLogContext(body string, contentType string) *gin.Context {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
	ctx.Request.Header.Set("Content-Type", contentType)
	ctx.Set("id", 7)
	ctx.Set("token_id", 42)
	ctx.Set("token_name", "Prod Key/One")
	ctx.Set("original_model", "gpt-4o")
	ctx.Set(common.RequestIdKey, "req-test")
	return ctx
}

func parseRequestBodyStdoutLogRecord(t *testing.T, output string) map[string]any {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	require.Len(t, lines, 1)
	require.True(t, strings.HasPrefix(lines[0], requestBodyStdoutLogPrefix))
	var record map[string]any
	require.NoError(t, common.Unmarshal([]byte(strings.TrimPrefix(lines[0], requestBodyStdoutLogPrefix)), &record))
	return record
}

func TestRecordRequestBodyStdoutLogSkipsWhenDisabled(t *testing.T) {
	withRequestBodyStdoutLogConfig(t, false, 100, 4096)
	buf := captureRequestBodyStdoutLog(t)
	ctx := newRequestBodyStdoutLogContext(`{"model":"gpt-4o","messages":[]}`, "application/json")

	RecordRequestBodyStdoutLog(ctx)

	require.Empty(t, buf.String())
}

func TestRecordRequestBodyStdoutLogSkipsWhenSamplePercentZero(t *testing.T) {
	withRequestBodyStdoutLogConfig(t, true, 0, 4096)
	buf := captureRequestBodyStdoutLog(t)
	ctx := newRequestBodyStdoutLogContext(`{"model":"gpt-4o","messages":[]}`, "application/json")

	RecordRequestBodyStdoutLog(ctx)

	require.Empty(t, buf.String())
}

func TestRecordRequestBodyStdoutLogWritesJSONPreview(t *testing.T) {
	body := "{\n  \"model\": \"gpt-4o\",\n  \"messages\": [{\"role\":\"user\",\"content\":\"hello\"}]\n}"
	withRequestBodyStdoutLogConfig(t, true, 100, 4096)
	buf := captureRequestBodyStdoutLog(t)
	ctx := newRequestBodyStdoutLogContext(body, "application/json")

	RecordRequestBodyStdoutLog(ctx)

	output := buf.String()
	require.Equal(t, 1, strings.Count(output, "\n"))
	record := parseRequestBodyStdoutLogRecord(t, output)
	require.Equal(t, "Prod Key/One", record["token_name"])
	require.Equal(t, "gpt-4o", record["model"])
	require.Equal(t, "/v1/chat/completions", record["path"])
	require.Equal(t, false, record["body_truncated"])
	preview, ok := record["body_preview"].(string)
	require.True(t, ok)
	require.NotContains(t, preview, "\n")
	require.Contains(t, preview, `"messages"`)
	require.Contains(t, preview, `"hello"`)

	storage, err := common.GetBodyStorage(ctx)
	require.NoError(t, err)
	replayed, err := storage.Bytes()
	require.NoError(t, err)
	require.Equal(t, body, string(replayed))
}

func TestRecordRequestBodyStdoutLogTruncatesLongLine(t *testing.T) {
	withRequestBodyStdoutLogConfig(t, true, 100, 512)
	buf := captureRequestBodyStdoutLog(t)
	ctx := newRequestBodyStdoutLogContext(`{"model":"gpt-4o","messages":[{"role":"user","content":"`+strings.Repeat("x", 4096)+`"}]}`, "application/json")

	RecordRequestBodyStdoutLog(ctx)

	line := strings.TrimSuffix(buf.String(), "\n")
	require.LessOrEqual(t, len(line), 512)
	record := parseRequestBodyStdoutLogRecord(t, buf.String())
	require.Equal(t, true, record["body_truncated"])
	preview, ok := record["body_preview"].(string)
	require.True(t, ok)
	require.NotEmpty(t, preview)
}

func TestRecordRequestBodyStdoutLogOmitsMultipartBody(t *testing.T) {
	withRequestBodyStdoutLogConfig(t, true, 100, 4096)
	buf := captureRequestBodyStdoutLog(t)
	ctx := newRequestBodyStdoutLogContext("file-binary-content", "multipart/form-data; boundary=test")

	RecordRequestBodyStdoutLog(ctx)

	record := parseRequestBodyStdoutLogRecord(t, buf.String())
	preview, ok := record["body_preview"].(string)
	require.True(t, ok)
	require.Contains(t, preview, "multipart/form-data body omitted")
	require.NotContains(t, preview, "file-binary-content")
}

func TestRecordRequestBodyStdoutLogConcurrentWritesCompleteLines(t *testing.T) {
	withRequestBodyStdoutLogConfig(t, true, 100, 4096)
	buf := captureRequestBodyStdoutLog(t)

	const requestCount = 20
	var wg sync.WaitGroup
	wg.Add(requestCount)
	for i := 0; i < requestCount; i++ {
		go func() {
			defer wg.Done()
			ctx := newRequestBodyStdoutLogContext(`{"model":"gpt-4o","messages":[{"role":"user","content":"hello"}]}`, "application/json")
			RecordRequestBodyStdoutLog(ctx)
		}()
	}
	wg.Wait()

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	require.Len(t, lines, requestCount)
	for _, line := range lines {
		require.True(t, strings.HasPrefix(line, requestBodyStdoutLogPrefix))
		var record map[string]any
		require.NoError(t, common.Unmarshal([]byte(strings.TrimPrefix(line, requestBodyStdoutLogPrefix)), &record))
		require.Equal(t, "Prod Key/One", record["token_name"])
	}
}
