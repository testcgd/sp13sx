package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"sp13sx/internal/config"
	"sp13sx/internal/util"
)

type Client struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewClient(cfg config.Backend) (*Client, error) {
	apiKey, ok := util.LookupEnv(cfg.APIKeyEnv)
	if !ok {
		return nil, fmt.Errorf("missing env var %s", cfg.APIKeyEnv)
	}
	return &Client{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:  apiKey,
		client:  &http.Client{},
	}, nil
}

func (c *Client) Post(ctx context.Context, path string, payload any) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		data, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("openai request failed: %s", strings.TrimSpace(string(data)))
	}
	return io.ReadAll(res.Body)
}

func (c *Client) Stream(ctx context.Context, path string, payload any) (<-chan []byte, <-chan error, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	res, err := c.client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	if res.StatusCode >= 300 {
		defer res.Body.Close()
		data, _ := io.ReadAll(res.Body)
		return nil, nil, fmt.Errorf("openai request failed: %s", strings.TrimSpace(string(data)))
	}

	out := make(chan []byte, 32)
	errs := make(chan error, 1)
	go func() {
		defer close(out)
		defer close(errs)
		defer res.Body.Close()

		scanner := bufio.NewScanner(res.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return
			}
			out <- []byte(data)
		}
		if err := scanner.Err(); err != nil {
			errs <- err
		}
	}()
	return out, errs, nil
}
