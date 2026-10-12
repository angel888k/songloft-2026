package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"songloft/internal/database"
)

const githubDiscoveryTokenConfigKey = "github_discovery_token"

type githubDiscoveryTokenSetting struct {
	Token string `json:"token"`
}

// GetGithubDiscoveryTokenSetting 获取客户端 GitHub 发现使用的凭据。
// @Summary 获取 GitHub 发现 Token
// @Description 返回服务器保存的完整 GitHub Token，供已登录客户端直接请求 GitHub API。未配置时 token 为空；服务器不代发发现请求。响应禁止缓存。
// @Tags JS插件管理
// @Produce json
// @Success 200 {object} githubDiscoveryTokenSetting "GitHub Token"
// @Failure 401 {object} models.ErrorResponse "未登录"
// @Failure 500 {object} models.ErrorResponse "读取配置失败"
// @Security BearerAuth
// @Router /settings/github-discovery-token [get]
func (h *JSPluginHandler) GetGithubDiscoveryTokenSetting(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	config, err := h.configService.GetConfig(r.Context(), githubDiscoveryTokenConfigKey)
	if errors.Is(err, database.ErrNotFound) {
		respondJSON(w, http.StatusOK, githubDiscoveryTokenSetting{})
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "读取 GitHub Token 失败", err)
		return
	}
	respondJSON(w, http.StatusOK, githubDiscoveryTokenSetting{Token: config.Value})
}

// UpdateGithubDiscoveryTokenSetting 保存或清除客户端 GitHub 发现使用的凭据。
// @Summary 保存 GitHub 发现 Token
// @Description 将 GitHub Token 保存到服务器配置，供各客户端共用。token 必填，空字符串表示清除；去掉首尾空白，内部只接受 ASCII 字母、数字和下划线，最长 512 字节。立即生效，无需预先创建配置；响应禁止缓存。
// @Tags JS插件管理
// @Accept json
// @Produce json
// @Param request body githubDiscoveryTokenSetting true "GitHub Token，空字符串表示清除"
// @Success 200 {object} githubDiscoveryTokenSetting "保存后的 Token"
// @Failure 400 {object} models.ErrorResponse "请求格式或 Token 格式错误"
// @Failure 401 {object} models.ErrorResponse "未登录"
// @Failure 500 {object} models.ErrorResponse "保存配置失败"
// @Security BearerAuth
// @Router /settings/github-discovery-token [put]
func (h *JSPluginHandler) UpdateGithubDiscoveryTokenSetting(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	var req struct {
		Token *string `json:"token"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	if err := decoder.Decode(&req); err != nil || req.Token == nil {
		respondError(w, http.StatusBadRequest, "请求必须包含 token 字符串", nil)
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		respondError(w, http.StatusBadRequest, "请求必须是单个 JSON 对象", nil)
		return
	}
	token := strings.TrimSpace(*req.Token)
	if len(token) > 512 || strings.IndexFunc(token, func(c rune) bool {
		return !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_')
	}) >= 0 {
		respondError(w, http.StatusBadRequest, "GitHub Token 格式错误", nil)
		return
	}
	if err := h.configService.Set(githubDiscoveryTokenConfigKey, token); err != nil {
		respondError(w, http.StatusInternalServerError, "保存 GitHub Token 失败", err)
		return
	}
	respondJSON(w, http.StatusOK, githubDiscoveryTokenSetting{Token: token})
}
