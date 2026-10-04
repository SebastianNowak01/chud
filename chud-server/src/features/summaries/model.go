package summaries

import (
	"time"

	"github.com/sebnow/chud/platform/clock"
)

type Status string

const (
	StatusReady       Status = "ready"
	StatusGenerating  Status = "generating"
	StatusUnavailable Status = "unavailable"
	StatusDisabled    Status = "disabled"
	StatusEmpty       Status = "empty"
)

type Summary struct {
	Week          clock.Date `db:"week"`
	PromptVersion int        `db:"prompt_version"`
	Text          string     `db:"text"`
	InputHash     string     `db:"input_hash"`
	GeneratedAt   time.Time  `db:"generated_at"`
	ExpiresAt     *time.Time `db:"expires_at"`
}

type WeekSummary struct {
	Week                  clock.Date `json:"week"`
	Status                Status     `json:"status"`
	Text                  string     `json:"text"`
	GeneratedAt           *time.Time `json:"generatedAt"`
	Outdated              *bool      `json:"outdated"`
	RegenerateAvailableAt *time.Time `json:"regenerateAvailableAt"`
}
