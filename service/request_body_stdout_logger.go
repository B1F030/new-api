package service

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"

	"github.com/gin-gonic/gin"
)

const (
	requestBodyStdoutLogPrefix               = "[LOG] "
	defaultRequestBodyLogMaxLineChars        = 4096
	multipartRequestBodyStdoutLogPlaceholder = "[multipart/form-data body omitted; size=%d]"
)

type requestBodyStdoutLogConfig struct {
	Enabled       bool
	SamplePercent int
	MaxLineChars  int
}

type requestBodyStdoutLogRecord struct {
	Time          int64  `json:"time"`
	RequestID     string `json:"request_id"`
	TokenID       int    `json:"token_id"`
	TokenName     string `json:"token_name"`
	Model         string `json:"model"`
	Path          string `json:"path"`
	Method        string `json:"method"`
	ContentType   string `json:"content_type"`
	BodySize      int64  `json:"body_size"`
	BodyTruncated bool   `json:"body_truncated"`
	BodyPreview   string `json:"body_preview"`
}

var (
	requestBodyStdoutLogMu     sync.Mutex
	requestBodyStdoutLogWriter io.Writer = os.Stdout
)

func RecordRequestBodyStdoutLog(c *gin.Context) {
	cfg := getRequestBodyStdoutLogConfig()
	if !cfg.Enabled || c == nil || c.Request == nil || c.Request.Body == nil || c.Request.Body == http.NoBody {
		return
	}
	if c.GetInt("token_id") <= 0 || !shouldSampleRequestBodyStdoutLog(cfg.SamplePercent) {
		return
	}

	line := buildRequestBodyStdoutLogLine(c, cfg)
	if len(line) == 0 {
		return
	}

	requestBodyStdoutLogMu.Lock()
	defer requestBodyStdoutLogMu.Unlock()
	_, _ = requestBodyStdoutLogWriter.Write(append(line, '\n'))
}

func getRequestBodyStdoutLogConfig() requestBodyStdoutLogConfig {
	samplePercent := constant.RequestBodyLogSamplePercent
	if samplePercent < 0 {
		samplePercent = 0
	}
	if samplePercent > 100 {
		samplePercent = 100
	}
	maxLineChars := constant.RequestBodyLogMaxLineChars
	if maxLineChars <= 0 {
		maxLineChars = defaultRequestBodyLogMaxLineChars
	}
	return requestBodyStdoutLogConfig{
		Enabled:       constant.RequestBodyLogEnabled,
		SamplePercent: samplePercent,
		MaxLineChars:  maxLineChars,
	}
}

func shouldSampleRequestBodyStdoutLog(samplePercent int) bool {
	if samplePercent <= 0 {
		return false
	}
	if samplePercent >= 100 {
		return true
	}
	return common.GetRandomInt(100) < samplePercent
}

func buildRequestBodyStdoutLogLine(c *gin.Context, cfg requestBodyStdoutLogConfig) []byte {
	storage, err := common.GetBodyStorage(c)
	if err != nil {
		return nil
	}
	defer resetRequestBodyForStdoutLog(c, storage)

	bodySize := storage.Size()
	if bodySize == 0 {
		return nil
	}

	contentType := c.Request.Header.Get("Content-Type")
	preview, err := buildRequestBodyPreview(storage, contentType)
	if err != nil {
		return nil
	}

	record := requestBodyStdoutLogRecord{
		Time:        common.GetTimestamp(),
		RequestID:   c.GetString(common.RequestIdKey),
		TokenID:     c.GetInt("token_id"),
		TokenName:   c.GetString("token_name"),
		Model:       c.GetString("original_model"),
		Path:        requestStdoutLogPath(c.Request),
		Method:      c.Request.Method,
		ContentType: contentType,
		BodySize:    bodySize,
		BodyPreview: preview,
	}
	return fitRequestBodyStdoutLogLine(record, cfg.MaxLineChars)
}

func buildRequestBodyPreview(storage common.BodyStorage, contentType string) (string, error) {
	normalizedContentType := strings.ToLower(contentType)
	if strings.HasPrefix(normalizedContentType, "multipart/form-data") {
		return fmt.Sprintf(multipartRequestBodyStdoutLogPlaceholder, storage.Size()), nil
	}

	bodyBytes, err := storage.Bytes()
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(normalizedContentType, "application/json") {
		var body any
		if err := common.Unmarshal(bodyBytes, &body); err == nil {
			if compactBody, marshalErr := common.Marshal(body); marshalErr == nil {
				return string(compactBody), nil
			}
		}
	}
	return strings.ToValidUTF8(string(bodyBytes), "?"), nil
}

func fitRequestBodyStdoutLogLine(record requestBodyStdoutLogRecord, maxLineChars int) []byte {
	if maxLineChars <= len(requestBodyStdoutLogPrefix) {
		return nil
	}
	line, ok := marshalRequestBodyStdoutLogLine(record, maxLineChars)
	if ok {
		return line
	}

	record.BodyTruncated = true
	preview := record.BodyPreview
	low, high := 0, len(preview)
	best := ""
	for low <= high {
		mid := (low + high) / 2
		record.BodyPreview = validUTF8Prefix(preview, mid)
		if _, ok := marshalRequestBodyStdoutLogLine(record, maxLineChars); ok {
			best = record.BodyPreview
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	record.BodyPreview = best
	line, ok = marshalRequestBodyStdoutLogLine(record, maxLineChars)
	if !ok {
		return nil
	}
	return line
}

func marshalRequestBodyStdoutLogLine(record requestBodyStdoutLogRecord, maxLineChars int) ([]byte, bool) {
	data, err := common.Marshal(record)
	if err != nil {
		return nil, false
	}
	line := make([]byte, 0, len(requestBodyStdoutLogPrefix)+len(data))
	line = append(line, requestBodyStdoutLogPrefix...)
	line = append(line, data...)
	return line, len(line) <= maxLineChars
}

func validUTF8Prefix(s string, maxBytes int) string {
	if maxBytes >= len(s) {
		return s
	}
	if maxBytes <= 0 {
		return ""
	}
	for maxBytes > 0 && !utf8.ValidString(s[:maxBytes]) {
		maxBytes--
	}
	return s[:maxBytes]
}

func resetRequestBodyForStdoutLog(c *gin.Context, storage common.BodyStorage) {
	if _, seekErr := storage.Seek(0, io.SeekStart); seekErr == nil {
		c.Request.Body = io.NopCloser(storage)
	}
}

func requestStdoutLogPath(req *http.Request) string {
	if req == nil || req.URL == nil {
		return ""
	}
	return req.URL.Path
}
