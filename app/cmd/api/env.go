package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

const (
	notionClientID     = "NOTION_CLIENT_ID"
	notionClientSecret = "NOTION_CLIENT_SECRET"
)

// NotionExtMode defines in which mode the credentials would be provided.
type NotionExtMode string

var (
	// Internal type of the extension (expects just the Secret Token to be provided).
	NotionExtModeInternal NotionExtMode = "internal"

	// Public expects the extension to be submitted for moderation and to be approved.
	// Will use OAuth2 authentication methods to talk to the Notion API.
	NotionExtModePublic NotionExtMode = "public"
)

var defaultNotionExtMode = NotionExtModeInternal

var validNotionExtModes = []NotionExtMode{
	NotionExtModeInternal,
	NotionExtModePublic,
}

// Validate the NotionExtMode.
func (m *NotionExtMode) Validate() error {
	if m == nil {
		return errors.New("notion ext mode is required")
	}
	for _, vm := range validNotionExtModes {
		if vm == *m {
			return nil
		}
	}

	return fmt.Errorf("unknown notion ext mode: %s", *m)
}

// Env with all the environment vars.
type Env struct {
	HTTP struct {
		Port string
	}

	GCloud struct {
		ProjectID   string
		ProjectName string
		LocationID  string
	}

	Segment struct {
		WriteKey string
	}

	Notion struct {
		ClientID     string
		ClientSecret string
		ExtMode      NotionExtMode

		// amount of workspaces processed in one go.
		ProcWss int64
	}
}

// NewEnv is a constructor for Env.
func NewEnv(ctx context.Context) (Env, error) {
	var e Env

	e.HTTP.Port = os.Getenv("PORT")
	if e.HTTP.Port == "" {
		e.HTTP.Port = "8080"
	}

	e.Segment.WriteKey = os.Getenv("SEGMENT_WRITE_KEY")

	e.GCloud.ProjectID = os.Getenv("GCLOUD_PROJECT_ID")
	e.GCloud.LocationID = os.Getenv("GCLOUD_LOCATION_ID")

	e.Notion.ExtMode = NotionExtMode(os.Getenv("NOTION_EXT_MODE"))
	if e.Notion.ExtMode == "" {
		e.Notion.ExtMode = defaultNotionExtMode
	}
	if err := e.Notion.ExtMode.Validate(); err != nil {
		return Env{}, err
	}

	procWssCount, err := strconv.ParseInt(os.Getenv("NOTION_PROC_WSS_COUNT"), 10, 64)
	if err != nil {
		log.Printf("Invalid value at NOTION_PROC_WSS_COUNT. Using default of 100")
		e.Notion.ProcWss = 100
	} else {
		e.Notion.ProcWss = procWssCount
	}

	if os.Getenv("ENV") != "appengine" {
		e.Notion.ClientID = os.Getenv(notionClientID)
		e.Notion.ClientSecret = os.Getenv(notionClientSecret)
	} else {
		client, err := secretmanager.NewClient(ctx)
		if err != nil {
			return Env{}, err
		}

		clientID, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
			Name: "projects/" + e.GCloud.ProjectID + "/secrets/" + notionClientID + "/versions/latest",
		})
		if err != nil {
			return Env{}, err
		}
		e.Notion.ClientID = string(clientID.GetPayload().GetData())

		clientSecret, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
			Name: "projects/" + e.GCloud.ProjectID + "/secrets/" + notionClientSecret + "/versions/latest",
		})
		if err != nil {
			return Env{}, err
		}
		e.Notion.ClientSecret = string(clientSecret.GetPayload().GetData())
	}

	return e, nil
}
