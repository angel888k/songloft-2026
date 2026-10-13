package app

import (
	"log/slog"
	"net/http"
	"time"

	chi_middleware "github.com/go-chi/chi/v5/middleware"
)

// slogLogFormatter 把 chi 的 access log 桥接到 slog，让访问日志受
// /settings/log-level 控制（默认 info 级别；调到 warn/error 时自动静默）。
//
// 替代 chi_middleware.Logger（其内部直接写标准库 log，不走 slog，导致
// 日志等级配置对它无效）。
type slogLogFormatter struct{}

func (slogLogFormatter) NewLogEntry(r *http.Request) chi_middleware.LogEntry {
	return &slogLogEntry{req: r, start: time.Now()}
}

type slogLogEntry struct {
	req   *http.Request
	start time.Time
}

func (e *slogLogEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, _ any) {
	attrs := []any{
		"method", e.req.Method,
		"path", e.req.URL.Path,
		"status", status,
		"bytes", bytes,
		"dur_ms", float64(elapsed.Microseconds()) / 1000.0,
		"remote", e.req.RemoteAddr,
	}

	// 记录字节区间与声明长度，区分分段读取、从头重拉及客户端提前断开。
	// 不记录 URL 查询串，其中可能包含播放器认证 token。
	for _, field := range []struct {
		name  string
		value string
	}{
		{"range", e.req.Header.Get("Range")},
		{"content_range", header.Get("Content-Range")},
		{"content_length", header.Get("Content-Length")},
	} {
		if field.value != "" {
			attrs = append(attrs, field.name, field.value)
		}
	}
	if reqID := chi_middleware.GetReqID(e.req.Context()); reqID != "" {
		attrs = append(attrs, "request_id", reqID)
	}
	slog.Info("access", attrs...)
}

func (e *slogLogEntry) Panic(v any, stack []byte) {
	slog.Error("access panic",
		"method", e.req.Method,
		"path", e.req.URL.Path,
		"panic", v,
		"stack", string(stack),
	)
}
