package embeddings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type LocalHTTP struct {
	BaseURL string
	HTTP    *http.Client
}

func NewLocalHTTP(baseURL string) *LocalHTTP {
	return &LocalHTTP{
		BaseURL: baseURL,
		HTTP:    &http.Client{Timeout: 120 * time.Second},
	}
}

type embedReq struct {
	Texts []string `json:"texts"`
}

type embedResp struct {
	Dim     int         `json:"dim"`
	Vectors [][]float64 `json:"vectors"`
}

func (e *LocalHTTP) Embed(text string) ([]float64, error) {
	vecs, err := e.EmbedBatch([]string{text})
	if err != nil {
		return nil, err
	}
	return vecs[0], nil
}

func (e *LocalHTTP) EmbedBatch(texts []string) ([][]float64, error) {
	b, _ := json.Marshal(embedReq{Texts: texts})
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/embed", e.BaseURL), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := e.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("embeddings service status=%s", resp.Status)
	}
	var out embedResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Vectors, nil
}

func (e *LocalHTTP) Dim() int {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/health", e.BaseURL), nil)
	resp, err := e.HTTP.Do(req)
	if err != nil {
		return 384
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return 384
	}
	var out struct {
		Dim int `json:"dim"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&out)
	if out.Dim == 0 {
		return 384
	}
	return out.Dim
}
