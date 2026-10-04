package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sebnow/chud/platform/config"
)

const (
	DefaultTimeout   = 3 * time.Minute
	maxResponseBytes = 1 << 20
	completionsPath  = "/v1/chat/completions"
)

var ErrDisabled = errors.New("llm is not configured")

type Config struct {
	URL     string
	Model   string
	Timeout time.Duration
}

func ConfigFromEnv() (Config, error) {
	cfg := Config{
		URL:     strings.TrimRight(strings.TrimSpace(os.Getenv(config.LLMURL)), "/"),
		Model:   strings.TrimSpace(os.Getenv(config.LLMModel)),
		Timeout: DefaultTimeout,
	}
	raw := strings.TrimSpace(os.Getenv(config.LLMTimeout))
	if raw == "" {
		return cfg, nil
	}
	timeout, err := time.ParseDuration(raw)
	if err != nil || timeout <= 0 {
		return cfg, fmt.Errorf("invalid %s %q, using %s", config.LLMTimeout, raw, DefaultTimeout)
	}
	cfg.Timeout = timeout
	return cfg, nil
}

type Completer interface {
	Enabled() bool
	Complete(ctx context.Context, system, prompt string) (string, error)
}

type Client struct {
	cfg  Config
	http *http.Client
}

func New(cfg Config) *Client {
	if cfg.Timeout <= 0 {
		cfg.Timeout = DefaultTimeout
	}
	return &Client{cfg: cfg, http: &http.Client{}}
}

func (c *Client) Enabled() bool {
	return c.cfg.URL != ""
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type completionRequest struct {
	Model    string    `json:"model,omitempty"`
	Messages []message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type completionResponse struct {
	Choices []struct {
		Message message `json:"message"`
	} `json:"choices"`
}

func (c *Client) Complete(ctx context.Context, system, prompt string) (string, error) {
	if !c.Enabled() {
		return "", ErrDisabled
	}
	ctx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	body, err := json.Marshal(completionRequest{
		Model: c.cfg.Model,
		Messages: []message{
			{Role: "system", Content: system},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.URL+completionsPath, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("llm request failed: %w", err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(io.LimitReader(res.Body, maxResponseBytes))
	if err != nil {
		return "", fmt.Errorf("llm response read failed: %w", err)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", fmt.Errorf("llm responded with %d: %s", res.StatusCode, strings.TrimSpace(string(data)))
	}

	var parsed completionResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", fmt.Errorf("llm response is not valid JSON: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", errors.New("llm response has no choices")
	}
	text := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if text == "" {
		return "", errors.New("llm response is empty")
	}
	return text, nil
}
