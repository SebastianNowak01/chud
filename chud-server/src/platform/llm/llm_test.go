package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sebnow/chud/platform/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigFromEnv(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		t.Setenv(config.LLMURL, "")
		t.Setenv(config.LLMTimeout, "")
		t.Setenv(config.LLMMaxTokens, "")
		cfg, err := ConfigFromEnv()
		require.NoError(t, err)
		assert.Equal(t, DefaultTimeout, cfg.Timeout)
		assert.Equal(t, DefaultMaxTokens, cfg.MaxTokens)
		assert.False(t, New(cfg).Enabled())
	})

	t.Run("custom values", func(t *testing.T) {
		t.Setenv(config.LLMURL, " http://localhost:11434/ ")
		t.Setenv(config.LLMModel, "qwen3:4b")
		t.Setenv(config.LLMTimeout, "90s")
		t.Setenv(config.LLMMaxTokens, "250")
		cfg, err := ConfigFromEnv()
		require.NoError(t, err)
		assert.Equal(t, Config{URL: "http://localhost:11434", Model: "qwen3:4b", Timeout: 90 * time.Second, MaxTokens: 250}, cfg)
		assert.True(t, New(cfg).Enabled())
	})

	t.Run("invalid values fall back to defaults", func(t *testing.T) {
		t.Setenv(config.LLMTimeout, "soon")
		t.Setenv(config.LLMMaxTokens, "lots")
		cfg, err := ConfigFromEnv()
		assert.ErrorContains(t, err, config.LLMTimeout)
		assert.ErrorContains(t, err, config.LLMMaxTokens)
		assert.Equal(t, DefaultTimeout, cfg.Timeout)
		assert.Equal(t, DefaultMaxTokens, cfg.MaxTokens)
	})
}

func TestComplete(t *testing.T) {
	respond := func(status int, body string) *Client {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(status)
			_, _ = w.Write([]byte(body))
		}))
		t.Cleanup(server.Close)
		return New(Config{URL: server.URL})
	}

	t.Run("sends messages and returns content", func(t *testing.T) {
		var got completionRequest
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, completionsPath, r.URL.Path)
			require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
			_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"  Dobry tydzień.  "}}]}`))
		}))
		defer server.Close()

		text, err := New(Config{URL: server.URL, Model: "m"}).Complete(context.Background(), "sys", "facts")
		require.NoError(t, err)
		assert.Equal(t, "Dobry tydzień.", text)
		assert.Equal(t, completionRequest{
			Model:     "m",
			Messages:  []message{{Role: "system", Content: "sys"}, {Role: "user", Content: "facts"}},
			MaxTokens: DefaultMaxTokens,
		}, got)
	})

	t.Run("disabled client", func(t *testing.T) {
		_, err := New(Config{}).Complete(context.Background(), "", "")
		assert.ErrorIs(t, err, ErrDisabled)
	})

	t.Run("unreachable server", func(t *testing.T) {
		server := httptest.NewServer(http.NotFoundHandler())
		url := server.URL
		server.Close()
		_, err := New(Config{URL: url}).Complete(context.Background(), "", "")
		assert.Error(t, err)
	})

	t.Run("timeout", func(t *testing.T) {
		release := make(chan struct{})
		server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			select {
			case <-release:
			case <-r.Context().Done():
			}
		}))
		defer server.Close()
		defer close(release)

		_, err := New(Config{URL: server.URL, Timeout: 50 * time.Millisecond}).Complete(context.Background(), "", "")
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("error status", func(t *testing.T) {
		_, err := respond(http.StatusServiceUnavailable, "loading model").Complete(context.Background(), "", "")
		assert.ErrorContains(t, err, "503")
	})

	t.Run("malformed body", func(t *testing.T) {
		_, err := respond(http.StatusOK, "not json").Complete(context.Background(), "", "")
		assert.Error(t, err)
	})

	t.Run("no choices", func(t *testing.T) {
		_, err := respond(http.StatusOK, `{"choices":[]}`).Complete(context.Background(), "", "")
		assert.Error(t, err)
	})

	t.Run("empty content", func(t *testing.T) {
		_, err := respond(http.StatusOK, `{"choices":[{"message":{"content":"  "}}]}`).Complete(context.Background(), "", "")
		assert.Error(t, err)
	})
}
