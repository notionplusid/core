package notion

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	autocounter "github.com/notionplusid/core/app"
)

type User struct {
	Object    string `json:"object"`
	ID        string `json:"id"`
	Type      string `json:"type,omitempty"`
	Name      string `json:"name,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Person    *struct {
		Email string `json:"email"`
	} `json:"person,omitempty"`
}

// Me returns the bot's User data.
// Primarily used to verify the access to the workspace.
func (n *Notion) Me(ctx context.Context) (User, error) {
	req, err := http.NewRequest(http.MethodGet, defaultAPIPath+"/v1/users/me", nil)
	if err != nil {
		return User{}, err
	}
	req.Header.Set("Notion-Version", defaultNotionAPIVersion)
	req.Header.Set("Authorization", "Bearer "+n.bearer)
	req.Header.Set("Content-Type", "application/json")

	res, err := n.http.Do(req.WithContext(ctx))
	if err != nil {
		return User{}, err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case 200:
	case 401:
		return User{}, autocounter.ErrUnauthorized
	default:
		e, _ := io.ReadAll(res.Body)
		return User{}, fmt.Errorf("unexpected error %d: %s", res.StatusCode, e)
	}

	var u User
	if err := json.NewDecoder(res.Body).Decode(&u); err != nil {
		return User{}, err
	}

	return u, nil
}
