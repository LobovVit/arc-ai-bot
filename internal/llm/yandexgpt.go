package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// YandexGPT implements the OpenAI-compatible Chat Completions API exposed by Yandex AI Studio.
// Docs:
// - OpenAI compatibility: base URL https://llm.api.cloud.yandex.net/v1 and header OpenAI-Project: <folder_ID>
// - Auth: API key can be passed as Bearer token in compatibility mode examples.
//
// We keep the request/response schema compatible with OpenAI's /chat/completions.
type YandexGPT struct {
	APIKey     string
	FolderID   string
	BaseURL    string
	Model      string
	AuthScheme string // "Bearer" or "Api-Key"
	HTTP       *http.Client
}

func NewYandexGPT(apiKey, folderID, baseURL, model, authScheme string) *YandexGPT {
	if baseURL == "" {
		baseURL = "https://llm.api.cloud.yandex.net/v1"
	}
	if authScheme == "" {
		authScheme = "Bearer"
	}
	return &YandexGPT{
		APIKey:     apiKey,
		FolderID:   folderID,
		BaseURL:    strings.TrimRight(baseURL, "/"),
		Model:      model,
		AuthScheme: authScheme,
		HTTP:       &http.Client{Timeout: 120 * time.Second},
	}
}

func (y *YandexGPT) Chat(system string, user string) (string, error) {
	if y.APIKey == "" {
		return "", fmt.Errorf("yandexgpt: api key is empty")
	}
	if y.FolderID == "" {
		return "", fmt.Errorf("yandexgpt: folder id is empty")
	}
	model := y.Model
	if model == "" {
		// Default per Yandex AI Studio OpenAI compatibility docs:
		// gpt://<folder_ID>/yandexgpt/latest
		model = fmt.Sprintf("gpt://%s/yandexgpt/latest", y.FolderID)
	}

	body := chatReq{
		Model: model,
		Messages: []chatMsg{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Temperature: 0.2,
	}
	b, _ := json.Marshal(body)

	url := y.BaseURL + "/chat/completions"
	req, _ := http.NewRequest("POST", url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	// OpenAI-compat requires folder to be passed as OpenAI-Project header.
	req.Header.Set("OpenAI-Project", y.FolderID)

	// Examples in Yandex docs use Bearer <API_key> in OpenAI-compat mode.
	// Official auth docs also mention Authorization: Api-Key <API_key> for some APIs.
	// We support both via YANDEX_AUTH_SCHEME.
	scheme := strings.TrimSpace(y.AuthScheme)
	if scheme == "" {
		scheme = "Bearer"
	}
	req.Header.Set("Authorization", scheme+" "+y.APIKey)

	resp, err := y.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("yandexgpt chat status=%s", resp.Status)
	}

	var out chatResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("yandexgpt chat: empty choices")
	}
	return out.Choices[0].Message.Content, nil
}
