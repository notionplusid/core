package autocounter

import (
	"errors"
	"time"
)

// Workspace domain structure.
type Workspace struct {
	ID          string    `json:"id"`
	Token       string    `json:"token"`
	ProcessedAt time.Time `json:"processedAt,omitempty"`
	CreatedAt   time.Time `json:"createdAt,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt,omitempty"`
}

// NewWorkspace constructor.
func NewWorkspace(id string, token string) (Workspace, error) {
	switch {
	case id == "":
		return Workspace{}, errors.New("workspace id is required")
	case token == "":
		return Workspace{}, errors.New("token is required")
	}

	now := time.Now()

	return Workspace{
		ID:          id,
		Token:       token,
		ProcessedAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Validate the Workspace.
func (ws *Workspace) Validate() error {
	switch {
	case ws.ID == "":
		return errors.New("id is required")
	case ws.Token == "":
		return errors.New("token is required")
	}

	return nil
}
