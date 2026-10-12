package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"songloft/internal/database/testutil"
	"songloft/internal/services"
)

func TestGithubDiscoveryTokenSettingPersistence(t *testing.T) {
	db := testutil.OpenMemoryDB(t)
	h := &JSPluginHandler{configService: services.NewConfigService(db.ConfigRepository())}
	read := func(handler *JSPluginHandler, expected string) {
		t.Helper()
		rr := httptest.NewRecorder()
		handler.GetGithubDiscoveryTokenSetting(rr, httptest.NewRequest(http.MethodGet, "/api/v1/settings/github-discovery-token", nil))
		if rr.Code != http.StatusOK || rr.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("GET status=%d headers=%v", rr.Code, rr.Header())
		}
		var setting githubDiscoveryTokenSetting
		if err := json.Unmarshal(rr.Body.Bytes(), &setting); err != nil || setting.Token != expected {
			t.Fatal("unexpected token setting")
		}
	}
	read(h, "")
	for _, token := range []string{"  github_pat_example_123  ", "ghp_replacement", ""} {
		body, _ := json.Marshal(githubDiscoveryTokenSetting{Token: token})
		rr := httptest.NewRecorder()
		h.UpdateGithubDiscoveryTokenSetting(rr, httptest.NewRequest(http.MethodPut, "/api/v1/settings/github-discovery-token", strings.NewReader(string(body))))
		if rr.Code != http.StatusOK || rr.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("PUT status=%d", rr.Code)
		}
		// Recreate the service to prove this lives in the database, not only its cache.
		read(&JSPluginHandler{configService: services.NewConfigService(db.ConfigRepository())}, strings.TrimSpace(token))
	}
}

func TestGithubDiscoveryTokenSettingRejectsInvalidRequests(t *testing.T) {
	db := testutil.OpenMemoryDB(t)
	h := &JSPluginHandler{configService: services.NewConfigService(db.ConfigRepository())}
	if err := h.configService.Set(githubDiscoveryTokenConfigKey, "ghp_original"); err != nil {
		t.Fatal(err)
	}
	for _, body := range []string{"invalid", "{}", `{"token":null}`, `{"token":23}`, `{"token":"bad\r\nheader"}`, `{"token":"bad token"}`, `{"token":"秘密"}`, `{"token":"ghp_valid"} {}`, `{"token":"` + strings.Repeat("a", 513) + `"}`, strings.Repeat(" ", 4097)} {
		rr := httptest.NewRecorder()
		h.UpdateGithubDiscoveryTokenSetting(rr, httptest.NewRequest(http.MethodPut, "/api/v1/settings/github-discovery-token", strings.NewReader(body)))
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
		if got := h.configService.GetString(githubDiscoveryTokenConfigKey, ""); got != "ghp_original" {
			t.Fatal("invalid request changed the saved credential")
		}
	}
}
