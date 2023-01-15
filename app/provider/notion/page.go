package notion

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	autocounter "github.com/notionplusid/core/app"
)

// Page object.
type Page struct {
	ID             string    `json:"id"`
	Object         string    `json:"object"`
	URL            string    `json:"url"`
	Archived       bool      `json:"archived"`
	CreatedTime    time.Time `json:"created_time"`
	LastEditedTime time.Time `json:"last_edited_at"`
	Parent         struct {
		Type       string `json:"type"`
		PageID     string `json:"page_id,omitempty"`
		Workspace  bool   `json:"workspace,omitempty"`
		DatabaseID string `json:"database_id,omitempty"`
	} `json:"parent"`
	Properties map[string]PageProperty `json:"properties"`
}

// PageProperty as expected in Page struct.
type PageProperty struct {
	ID       string       `json:"id,omitempty"`
	Type     PropertyType `json:"type"`
	Title    []RichText   `json:"title,omitempty"`
	RichText []RichText   `json:"rich_text,omitempty"`
	Number   *float64     `json:"number,omitempty"`
	Formula  *struct {
		Type    string            `json:"type"`
		String  *string           `json:"string,omitempty"`
		Number  *float64          `json:"number,omitempty"`
		Boolean *bool             `json:"boolean,omitempty"`
		Date    *PagePropertyDate `json:"date,omitempty"`
	} `json:"formula,omitempty"`
	Relation []struct {
		ID string `json:"id"`
	} `json:"relation,omitempty"`
	Rollup *struct {
		Type   string            `json:"type"`
		Number int64             `json:"number,omitempty"`
		Date   *PagePropertyDate `json:"date,omitempty"`
		Array  []PageProperty    `json:"array,omitempty"`
	} `json:"rollup,omitempty"`
	Select *struct {
		ID    string `json:"id,omitempty"`
		Name  string `json:"name,omitempty"`
		Color string `json:"color,omitempty"`
	} `json:"select,omitempty"`
	MultiSelect []struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Color string `json:"color"`
	} `json:"multi_select,omitempty"`
	Date   *PagePropertyDate `json:"date,omitempty"`
	People []User            `json:"people,omitempty"`
	Files  []struct {
		Name string `json:"name"`
	} `json:"files,omitempty"`
	Checkbox       *bool      `json:"checkbox,omitempty"`
	URL            *string    `json:"url,omitempty"`
	Email          *string    `json:"email,omitempty"`
	PhoneNumber    *string    `json:"phone_number,omitempty"`
	CreatedTime    *time.Time `json:"created_time,omitempty"`
	CreatedBy      *User      `json:"created_by,omitempty"`
	LastEditedTime *time.Time `json:"last_edited_time,omitempty"`
	LastEditedBy   *User      `json:"last_edited_by,omitempty"`
}

// PagePropertyDate format allows to cast the otherwise string formatted date property values.
type PagePropertyDate struct {
	// "2006-01-02"
	Start time.Time `json:"start"`
	End   time.Time `json:"end,omitempty"`
}

// MarshalJSON implementation.
func (p *PagePropertyDate) MarshalJSON() ([]byte, error) {
	if p == nil {
		return nil, nil
	}

	pRaw := struct {
		Start string `json:"start"`
		End   string `json:"end,omitempty"`
	}{}
	pRaw.Start = p.Start.Format("2006-01-02")
	if !p.End.IsZero() {
		pRaw.End = p.End.Format("2006-01-02")
	}

	return json.Marshal(pRaw)
}

// UnmarshalJSON implementation.
func (p *PagePropertyDate) UnmarshalJSON(j []byte) error {
	if p == nil {
		return nil
	}
	if len(j) == 0 {
		return nil
	}

	pRaw := struct {
		Start string `json:"start"`
		End   string `json:"end,omitempty"`
	}{}
	if err := json.Unmarshal(j, &pRaw); err != nil {
		return nil
	}

	pStart, err := time.Parse("2006-01-02", pRaw.Start)
	if err != nil {
		return fmt.Errorf("couldn't parse Start: %w", err)
	}
	p.Start = pStart

	if pRaw.End == "" {
		return nil
	}

	pEnd, err := time.Parse("2006-01-02", pRaw.End)
	if err != nil {
		return fmt.Errorf("couldn't parse End: %w", err)
	}
	p.End = pEnd

	return nil
}

type PatchPageReq struct {
	Archived   bool                    `json:"archived,omitempty"`
	Properties map[string]PageProperty `json:"properties"`
}

// PatchPage with provided ID and data.
func (n *Notion) PatchPage(ctx context.Context, pageID string, q PatchPageReq) (Page, error) {
	if pageID == "" {
		return Page{}, errors.New("page id is required")
	}

	b, err := json.Marshal(&q)
	if err != nil {
		return Page{}, fmt.Errorf("couldn't parse patch page request: %s", err)
	}

	req, err := http.NewRequest(http.MethodPatch, defaultAPIPath+"/v1/pages/"+pageID, bytes.NewBuffer(b))
	if err != nil {
		return Page{}, err
	}
	req.Header.Set("Notion-Version", defaultNotionAPIVersion)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+n.bearer)

	res, err := n.http.Do(req.WithContext(ctx))
	if err != nil {
		return Page{}, err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case 200:
	case 404:
		return Page{}, autocounter.ErrPageNotFound
	default:
		e, _ := io.ReadAll(res.Body)
		return Page{}, fmt.Errorf("unexpected error %d: %s", res.StatusCode, e)
	}

	var qRes Page
	if err := json.NewDecoder(res.Body).Decode(&qRes); err != nil {
		return Page{}, err
	}

	return qRes, nil
}
