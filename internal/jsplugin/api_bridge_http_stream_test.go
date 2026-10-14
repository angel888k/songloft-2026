package jsplugin

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func openTestHTTPStream(t *testing.T, h *BridgeHandler, url string, timeout int) string {
	t.Helper()
	params, _ := json.Marshal(map[string]any{"url": url, "timeoutMs": timeout})
	raw, err := h.HandleBridgeCall("net.httpStreamCreate", string(params))
	if err != nil {
		t.Fatal(err)
	}
	var value struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(`{"id":%q}`, value.ID)
}

func TestHTTPStreamReadsBeforeResponseFinishesAndCancels(t *testing.T) {
	closed := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Set-Cookie", "a=1")
		w.Header().Add("Set-Cookie", "b=2")
		w.Write([]byte("首段"))
		w.(http.Flusher).Flush()
		<-r.Context().Done()
		close(closed)
	}))
	defer server.Close()
	h := newTestNetBridgeHandler(t)
	t.Cleanup(h.cleanupHTTPStreams)
	h.service.bridgeHandler = h
	id := openTestHTTPStream(t, h, server.URL, 30000)
	if !h.service.HasActiveHTTPStreams() {
		t.Fatal("active stream not visible to idle checks")
	}
	headers, err := h.HandleBridgeCall("net.httpStreamHeaders", id)
	if err != nil {
		t.Fatal(err)
	}
	var response struct {
		Headers map[string][]string `json:"headers"`
	}
	json.Unmarshal([]byte(headers), &response)
	if len(response.Headers["Set-Cookie"]) != 2 {
		t.Fatalf("lost headers: %s", headers)
	}
	raw, err := h.HandleBridgeCall("net.httpStreamRead", id)
	if err != nil {
		t.Fatal(err)
	}
	var chunk struct {
		Data string `json:"data"`
	}
	json.Unmarshal([]byte(raw), &chunk)
	bytes, _ := base64.StdEncoding.DecodeString(chunk.Data)
	if string(bytes) != "首段" {
		t.Fatalf("chunk: %s", raw)
	}
	pending := make(chan error, 1)
	go func() { _, err := h.HandleBridgeCall("net.httpStreamRead", id); pending <- err }()
	h.HandleBridgeCall("net.httpStreamCancel", id)
	select {
	case err := <-pending:
		if err == nil {
			t.Fatal("read survived cancellation")
		}
	case <-time.After(time.Second):
		t.Fatal("read not cancelled")
	}
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("upstream not cancelled")
	}
	if _, err := h.HandleBridgeCall("net.httpStreamCancel", id); err != nil {
		t.Fatal(err)
	}
	if h.service.HasActiveHTTPStreams() {
		t.Fatal("cancelled stream keeps plugin awake")
	}
}

func TestHTTPStreamDeadlineAndCleanupDuringHeaders(t *testing.T) {
	for _, cleanup := range []bool{false, true} {
		t.Run(fmt.Sprint(cleanup), func(t *testing.T) {
			started, closed := make(chan struct{}), make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { close(started); <-r.Context().Done(); close(closed) }))
			defer server.Close()
			h := newTestNetBridgeHandler(t)
			t.Cleanup(h.cleanupHTTPStreams)
			id := openTestHTTPStream(t, h, server.URL, 150)
			<-started
			if cleanup {
				h.cleanupHTTPStreams()
			}
			if _, err := h.HandleBridgeCall("net.httpStreamHeaders", id); err == nil {
				t.Fatal("headers succeeded after timeout/cleanup")
			}
			select {
			case <-closed:
			case <-time.After(time.Second):
				t.Fatal("upstream not closed")
			}
			if cleanup {
				if _, err := h.HandleBridgeCall("net.httpStreamCreate", fmt.Sprintf(`{"url":%q}`, server.URL)); err == nil {
					t.Fatal("creation after cleanup")
				}
			}
		})
	}
}

func TestHTTPStreamPermissionLimitAndEOF(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("abc")) }))
	defer server.Close()
	h := newTestNetBridgeHandler(t)
	t.Cleanup(h.cleanupHTTPStreams)
	h.permissions = nil
	if _, err := h.HandleBridgeCall("net.httpStreamCreate", fmt.Sprintf(`{"url":%q}`, server.URL)); err == nil {
		t.Fatal("net permission not enforced")
	}
	h.permissions = []string{PermNet}
	ids := make([]string, 8)
	for i := range ids {
		ids[i] = openTestHTTPStream(t, h, server.URL, 30000)
	}
	if _, err := h.HandleBridgeCall("net.httpStreamCreate", fmt.Sprintf(`{"url":%q}`, server.URL)); err == nil {
		t.Fatal("limit not enforced")
	}
	for _, id := range ids {
		for {
			raw, err := h.HandleBridgeCall("net.httpStreamRead", id)
			if err != nil {
				t.Fatal(err)
			}
			var chunk struct {
				Done bool `json:"done"`
			}
			json.Unmarshal([]byte(raw), &chunk)
			if chunk.Done {
				break
			}
		}
	}
	h.httpMu.Lock()
	count := len(h.httpStreams)
	h.httpMu.Unlock()
	if count != 0 {
		t.Fatalf("EOF retained %d streams", count)
	}
}

func TestHTTPStreamQuickJSIncrementalRead(t *testing.T) {
	finished := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("first"))
		w.(http.Flusher).Flush()
		<-r.Context().Done()
		close(finished)
	}))
	defer server.Close()
	_, router := loadNetIntegrationTestPluginWithCode(t, "http-stream-test", fmt.Sprintf(`
globalThis.onHTTPRequest = async function() {
  var request = songloft.net.httpStream(%q);
  var headers = await request.response;
  var first = await request.read();
  await request.cancel();
  return {statusCode: 200, headers: {}, body: JSON.stringify({status: headers.status, first: Array.from(first)})};
};`, server.URL))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("GET", "/api/v1/jsplugin/http-stream-test/api/probe", nil))
	if w.Code != 200 || w.Body.String() != `{"status":200,"first":[102,105,114,115,116]}` {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("QuickJS cancellation did not reach upstream")
	}
}
