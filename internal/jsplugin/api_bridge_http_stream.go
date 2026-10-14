package jsplugin

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"songloft/internal/jsruntime"
)

type managedHTTPStream struct {
	stream *jsruntime.HTTPStream
	cancel context.CancelFunc
}

func (h *BridgeHandler) handleHTTPStream(action, data string) (string, error) {
	if action == "net.httpStreamCreate" {
		return h.createHTTPStream(data)
	}
	var params struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(data), &params); err != nil {
		return "", err
	}
	h.httpMu.Lock()
	entry := h.httpStreams[params.ID]
	h.httpMu.Unlock()
	if action == "net.httpStreamCancel" {
		h.removeHTTPStream(params.ID)
		return "", nil
	}
	if entry == nil {
		return "", fmt.Errorf("HTTP stream closed or expired")
	}
	if action == "net.httpStreamHeaders" {
		status, headers, err := entry.stream.Headers()
		if err != nil {
			h.removeHTTPStream(params.ID)
			return "", err
		}
		encoded, _ := json.Marshal(map[string]any{"status": status, "headers": headers})
		return string(encoded), nil
	}
	chunk, err := entry.stream.Read()
	if err != nil {
		h.removeHTTPStream(params.ID)
		return "", err
	}
	if chunk == nil {
		h.removeHTTPStream(params.ID)
		return `{"done":true}`, nil
	}
	encoded, _ := json.Marshal(map[string]any{"done": false, "data": base64.StdEncoding.EncodeToString(chunk)})
	return string(encoded), nil
}

func (h *BridgeHandler) createHTTPStream(data string) (string, error) {
	var params struct {
		URL       string            `json:"url"`
		Method    string            `json:"method"`
		Headers   map[string]string `json:"headers"`
		Body      string            `json:"body"`
		TimeoutMS int               `json:"timeoutMs"`
	}
	if err := json.Unmarshal([]byte(data), &params); err != nil {
		return "", err
	}
	u, err := url.Parse(params.URL)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return "", fmt.Errorf("HTTP stream requires an HTTP(S) URL")
	}
	if len(params.Body) > 4<<20 {
		return "", fmt.Errorf("HTTP stream request body exceeds 4 MiB")
	}
	if params.Method == "" {
		params.Method = "GET"
	}
	if params.TimeoutMS == 0 {
		params.TimeoutMS = 30000
	}
	if params.TimeoutMS < 100 || params.TimeoutMS > 30000 {
		return "", fmt.Errorf("timeoutMs must be 100–30000")
	}
	if params.Headers == nil {
		params.Headers = map[string]string{}
	}
	for key := range params.Headers {
		if strings.EqualFold(key, "X-Fetch-Timeout-Ms") {
			delete(params.Headers, key)
		}
	}
	params.Headers["X-Fetch-Timeout-Ms"] = strconv.Itoa(params.TimeoutMS)
	headers, _ := json.Marshal(params.Headers)
	h.httpMu.Lock()
	defer h.httpMu.Unlock()
	if h.httpClosed {
		return "", fmt.Errorf("plugin is stopping")
	}
	if len(h.httpStreams) >= 8 {
		return "", fmt.Errorf("max 8 HTTP streams per plugin")
	}
	if h.httpStreams == nil {
		h.httpStreams = make(map[string]*managedHTTPStream)
	}
	id := fmt.Sprintf("http-%d", h.socketIDSeq.Add(1))
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(params.TimeoutMS)*time.Millisecond)
	stream := jsruntime.NewHTTPStream(ctx, params.URL, params.Method, string(headers), hex.EncodeToString([]byte(params.Body)), CheckPermission(h.permissions, PermNetInsecureTLS))
	h.httpStreams[id] = &managedHTTPStream{stream: stream, cancel: cancel}
	context.AfterFunc(ctx, func() { h.removeHTTPStream(id) })
	encoded, _ := json.Marshal(map[string]string{"id": id})
	return string(encoded), nil
}

func (h *BridgeHandler) removeHTTPStream(id string) {
	h.httpMu.Lock()
	entry := h.httpStreams[id]
	delete(h.httpStreams, id)
	h.httpMu.Unlock()
	if entry != nil {
		entry.cancel()
		entry.stream.Close()
	}
}

func (h *BridgeHandler) cleanupHTTPStreams() {
	h.httpMu.Lock()
	h.httpClosed = true
	entries := h.httpStreams
	h.httpStreams = nil
	h.httpMu.Unlock()
	for _, entry := range entries {
		entry.cancel()
		entry.stream.Close()
	}
}
