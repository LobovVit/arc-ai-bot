package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// OpenAI implements a minimal Chat Completions client.
// We use it only for answer generation; embeddings remain local.
type OpenAI struct {
	APIKey string
	Model  string
	HTTP   *http.Client
}

func NewOpenAI(apiKey, model string) *OpenAI {
	return &OpenAI{
		APIKey: apiKey,
		Model:  model,
		HTTP:   &http.Client{Timeout: 120 * time.Second},
	}
}

type chatReq struct {
	Model       string    `json:"model"`
	Messages    []chatMsg `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
}

type chatMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResp struct {
	Choices []struct {
		Message chatMsg `json:"message"`
	} `json:"choices"`
}

func (o *OpenAI) Chat(system string, user string) (string, error) {
	body := chatReq{
		Model: o.Model,
		Messages: []chatMsg{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Temperature: 0.2,
	}
	b, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+o.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("openai chat status=%s", resp.Status)
	}

	var out chatResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("openai chat: empty choices")
	}
	return out.Choices[0].Message.Content, nil
}
