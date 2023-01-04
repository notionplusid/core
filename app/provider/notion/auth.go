package notion

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var client = &http.Client{}

// OAuth2Res is the Notion API response with the expected data.
type OAuth2Res struct {
	AccessToken          string          `json:"access_token"`
	BotID                string          `json:"bot_id"`
	DuplicatedTemplateID string          `json:"duplicated_template_id,omitempty"`
	Owner                json.RawMessage `json:"owner"`
	WorkspaceIconURL     string          `json:"workspace_icon,omitempty"`
	WorkspaceID          string          `json:"workspace_id"`
	WorkspaceName        string          `json:"workspace_name,omitempty"`
}

// ExtConfig for the Notion Extension.
type ExtConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// Validate the ExtConfig.
func (e *ExtConfig) Validate() error {
	switch {
	case e.ClientID == "":
		return errors.New("client id is required")
	case e.ClientSecret == "":
		return errors.New("client secret is required")
	case e.RedirectURI == "":
		return errors.New("redirect uri is required")
	}

	return nil
}

// OAuth2 authorises the Workspace with provided code and client ID.
func OAuth2(ctx context.Context, code string, config ExtConfig) (OAuth2Res, error) {
	if code == "" {
		return OAuth2Res{}, errors.New("code is required")
	}
	if err := config.Validate(); err != nil {
		return OAuth2Res{}, fmt.Errorf("invalid config: %w", err)
	}

	type reqBody struct {
		GrantType   string `json:"grant_type"`
		Code        string `json:"code"`
		RedirectURI string `json:"redirect_uri,omitempty"`
	}

	bs, err := json.Marshal(&reqBody{
		GrantType:   "authorization_code",
		Code:        code,
		RedirectURI: config.RedirectURI,
	})
	if err != nil {
		return OAuth2Res{}, err
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.notion.com/v1/oauth/token", bytes.NewReader(bs))
	if err != nil {
		return OAuth2Res{}, err
	}
	req.SetBasicAuth(config.ClientID, config.ClientSecret)
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return OAuth2Res{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		b, _ := io.ReadAll(res.Body)
		return OAuth2Res{}, fmt.Errorf("unexpected response %d: %s", res.StatusCode, b)
	}

	var resBody OAuth2Res
	if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
		return OAuth2Res{}, err
	}

	return resBody, nil
}
