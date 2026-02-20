package qdrant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

func New(baseURL string) *Client {
	return &Client{BaseURL: baseURL, HTTP: &http.Client{Timeout: 60 * time.Second}}
}

type CreateCollectionRequest struct {
	Vectors struct {
		Size     int    `json:"size"`
		Distance string `json:"distance"`
	} `json:"vectors"`
}

func (c *Client) CreateCollection(name string, vectorSize int) error {
	payload := CreateCollectionRequest{}
	payload.Vectors.Size = vectorSize
	payload.Vectors.Distance = "Cosine"
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/collections/%s", c.BaseURL, name), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 409 {
		return nil
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("create collection status=%s", resp.Status)
	}
	return nil
}

type Point struct {
	ID      interface{}            `json:"id"`
	Vector  []float64              `json:"vector"`
	Payload map[string]interface{} `json:"payload"`
}

type UpsertRequest struct {
	Points []Point `json:"points"`
}

func (c *Client) Upsert(collection string, points []Point) error {
	b, _ := json.Marshal(UpsertRequest{Points: points})
	req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/collections/%s/points?wait=true", c.BaseURL, collection), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("upsert status=%s", resp.Status)
	}
	return nil
}

type SearchRequest struct {
	Vector      []float64 `json:"vector"`
	Limit       int       `json:"limit"`
	WithPayload bool      `json:"with_payload"`
}

type SearchResult struct {
	Result []struct {
		Score   float64                `json:"score"`
		Payload map[string]interface{} `json:"payload"`
	} `json:"result"`
}

func (c *Client) Search(collection string, vector []float64, limit int) ([]map[string]interface{}, error) {
	b, _ := json.Marshal(SearchRequest{Vector: vector, Limit: limit, WithPayload: true})
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/collections/%s/points/search", c.BaseURL, collection), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("search status=%s", resp.Status)
	}
	var out SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	items := make([]map[string]interface{}, 0, len(out.Result))
	for _, r := range out.Result {
		p := r.Payload
		p["_score"] = r.Score
		items = append(items, p)
	}
	return items, nil
}
