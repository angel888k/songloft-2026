package jsplugin

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

const webSocketDuringHTTPRequestJSCode = `
var hostEvents = [];
var activeSocket;
var resolveHTTP;

globalThis.onInit = async function() {};
globalThis.onDeinit = async function() {};
globalThis.onHTTPRequest = async function(req) {
    hostEvents = [];
    var closed = new Promise(function(resolve) { resolveHTTP = resolve; });
    await activeSocket.send("http-started");
    await closed;
    return {
        statusCode: 200,
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ events: hostEvents })
    };
};
globalThis.onWebSocket = async function(req, socket) {
    activeSocket = socket;
    socket.onMessage(async function(event) {
        hostEvents.push("message:" + (event.isBinary ? "binary" : String(event.data)));
        if (!event.isBinary) {
            await socket.send("ack:" + event.data);
        }
    });
    socket.onClose(async function(event) {
        hostEvents.push("close:" + event.code + ":" + (event.reason || ""));
        if (resolveHTTP) resolveHTTP();
    });
    await socket.send("socket-ready");
};
`

func loadWebSocketIntegrationTestPlugin(t *testing.T, entryPath string) (*Manager, http.Handler) {
	t.Helper()

	pluginsDir, dataDir, repo, _ := setupTestEnv(t)
	ctx := context.Background()

	manifest := testManifest(entryPath)
	manifest.Permissions = []string{PermWebSocket}
	zipData := createTestPluginZip(t, manifest, webSocketDuringHTTPRequestJSCode)

	zipFileName := entryPath + ".jsplugin.zip"
	if err := os.WriteFile(filepath.Join(pluginsDir, zipFileName), zipData, 0o644); err != nil {
		t.Fatalf("write zip: %v", err)
	}

	plugin := &JSPlugin{
		Name:        manifest.Name,
		Version:     manifest.Version,
		Description: manifest.Description,
		Author:      manifest.Author,
		EntryPath:   manifest.EntryPath,
		Main:        manifest.Main,
		Permissions: manifest.Permissions,
		Status:      JSPluginStatusActive,
		FilePath:    zipFileName,
	}
	if err := repo.Create(ctx, plugin); err != nil {
		t.Fatalf("create plugin record: %v", err)
	}

	manager := NewManager(repo, pluginsDir, dataDir, "", nil, nil)
	t.Cleanup(func() { manager.Close() })

	if err := manager.LoadPlugin(ctx, plugin); err != nil {
		t.Skipf("LoadPlugin failed (may need QuickJS runtime): %v", err)
	}

	router := chi.NewRouter()
	manager.RegisterAPIRoutes(router)
	return manager, router
}

func TestInboundWebSocketHostEventDeliveredDuringHTTPRequest(t *testing.T) {
	_, router := loadWebSocketIntegrationTestPlugin(t, "ws-during-http")
	server := httptest.NewServer(router)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/v1/jsplugin/ws-during-http/api/inbound"
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		if resp != nil {
			t.Fatalf("dial websocket: status=%d err=%v", resp.StatusCode, err)
		}
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()
	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("set websocket read deadline: %v", err)
	}
	expectMessage := func(want string) {
		t.Helper()
		messageType, payload, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("read %q: %v", want, err)
		}
		if messageType != websocket.TextMessage || string(payload) != want {
			t.Fatalf("message = type %d payload %q, want text %q", messageType, string(payload), want)
		}
	}
	// The upgrade completes before the JS open handler registers callbacks.
	expectMessage("socket-ready")

	type httpResult struct {
		status int
		body   []byte
		err    error
	}
	resultCh := make(chan httpResult, 1)
	requestCtx, cancelRequest := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelRequest()
	request, err := http.NewRequestWithContext(requestCtx, http.MethodGet, server.URL+"/api/v1/jsplugin/ws-during-http/api/scan", nil)
	if err != nil {
		t.Fatalf("create plugin request: %v", err)
	}
	go func() {
		resp, err := http.DefaultClient.Do(request)
		if err != nil {
			resultCh <- httpResult{err: err}
			return
		}
		defer resp.Body.Close()
		body, readErr := io.ReadAll(resp.Body)
		resultCh <- httpResult{status: resp.StatusCode, body: body, err: readErr}
	}()

	// Wait for the HTTP handler to reset its event list and begin awaiting
	// close. A fixed sleep can send ping before that reset and lose the events.
	expectMessage("http-started")

	if err := conn.WriteMessage(websocket.TextMessage, []byte("ping")); err != nil {
		t.Fatalf("write text: %v", err)
	}
	expectMessage("ack:ping")

	if err := conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"),
		time.Now().Add(time.Second),
	); err != nil {
		t.Fatalf("write close: %v", err)
	}

	result := <-resultCh
	if result.err != nil {
		t.Fatalf("GET plugin route: %v", result.err)
	}
	if result.status != http.StatusOK {
		t.Fatalf("status = %d, body = %s", result.status, string(result.body))
	}

	var payloadBody struct {
		Events []string `json:"events"`
	}
	if err := json.Unmarshal(result.body, &payloadBody); err != nil {
		t.Fatalf("unmarshal body %s: %v", string(result.body), err)
	}
	if len(payloadBody.Events) < 2 {
		t.Fatalf("events = %#v, want message and close events", payloadBody.Events)
	}
	if payloadBody.Events[0] != "message:ping" {
		t.Fatalf("events[0] = %q, want %q", payloadBody.Events[0], "message:ping")
	}
	if payloadBody.Events[1] != "close:1000:bye" {
		t.Fatalf("events[1] = %q, want %q", payloadBody.Events[1], "close:1000:bye")
	}
}
