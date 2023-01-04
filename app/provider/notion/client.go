package notion

import (
	"errors"
	"fmt"
	"net/http"

	autocounter "github.com/notionplusid/core/main"
)

const defaultNotionAPIVersion = "2022-06-28"
const defaultAPIPath = "https://api.notion.com"

// Notion API client.
type Notion struct {
	bearer string
	http   *http.Client
}

// NewClient for Notion API.
func NewClient(bearerToken string) (*Notion, error) {
	if bearerToken == "" {
		return nil, errors.New("bearer token is required")
	}

	return &Notion{
		bearer: bearerToken,
		http:   &http.Client{},
	}, nil
}

// NewFromWorkspace initialiases Notion API client from the provided Workspace.
func NewFromWorkspace(ws autocounter.Workspace) (*Notion, error) {
	if err := ws.Validate(); err != nil {
		return nil, fmt.Errorf("workspace: %s", err)
	}

	return NewClient(ws.Token)
}

// WithHTTPClient returns a new Notion instance with the provided property assigned as a HTTP Client
// and uses for all the requests to the Notion API.
// If client is nil - the new empty http Client will be used instead.
func (n *Notion) WithHTTPClient(hc *http.Client) *Notion {
	copy := *n
	if hc == nil {
		return &copy
	}
	copy.http = hc
	return nil
}
