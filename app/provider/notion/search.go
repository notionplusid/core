package notion

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	autocounter "github.com/notionplusid/core/app"
)

const maxSearchPageSize = 100

// Known SortDirection values.
const (
	SortDirectionAsc  = "ascending"
	SortDirectionDesc = "descending"
)

var validSortDirections = []SortDirection{
	SortDirectionAsc,
	SortDirectionDesc,
}

// SortDirection as expected to be passed in SearchRequest.
type SortDirection string

// Validate the SortDirection.
func (sd SortDirection) Validate() error {
	if sd == "" {
		return nil
	}

	for _, val := range validSortDirections {
		if val == sd {
			return nil
		}
	}

	return fmt.Errorf("unknown sort direction: %s", sd)
}

// Known SortTimestampLastEdited values.
const (
	SortTimestampLastEdited = "last_edited_time"
)

var validSortTimestamps = []SortTimestamp{
	SortTimestampLastEdited,
}

// SortTimestamp as expected to be passed in SearchRequest.
type SortTimestamp string

// Validate the SortTimestamp.
func (st SortTimestamp) Validate() error {
	if st == "" {
		return nil
	}

	for _, val := range validSortTimestamps {
		if val == st {
			return nil
		}
	}

	return fmt.Errorf("unknown sort timestamp: %s", st)
}

// Sort configuration for SearchReq.
type Sort struct {
	Direction SortDirection `json:"direction,omitempty"`
	Timestamp SortTimestamp `json:"timestamp,omitempty"`
}

// Validate the Sort.
func (s *Sort) Validate() error {
	if s == nil {
		return nil
	}

	if err := s.Direction.Validate(); err != nil {
		return fmt.Errorf("direction: %s", err)
	}

	if err := s.Timestamp.Validate(); err != nil {
		return fmt.Errorf("timestamp: %s", err)
	}

	return nil
}

// Known FilterProp values.
const (
	FilterPropObject = "object"
)

var validFilterPropValues = []FilterProp{FilterPropObject}

// FilterProp as expected by the Filter.
type FilterProp string

// Validate the FilterProp.
func (fp FilterProp) Validate() error {
	if fp == "" {
		return nil
	}

	for _, val := range validFilterPropValues {
		if val == fp {
			return nil
		}
	}

	return fmt.Errorf("unknown filter property: %s", fp)
}

// Known FilterValues.
const (
	FilterValuePage = "page"
	FilterValueDB   = "database"
)

var validFilterValues = []FilterValue{FilterValuePage, FilterValueDB}

// FilterValue as expected in Filter.
type FilterValue string

// Validate the FilterValue.
func (fv FilterValue) Validate() error {
	if fv == "" {
		return nil
	}

	for _, val := range validFilterValues {
		if val == fv {
			return nil
		}
	}

	return fmt.Errorf("unknown filter value: %s", fv)
}

// Filter as expected in SearchReq.
type Filter struct {
	Prop  FilterProp  `json:"property"`
	Value FilterValue `json:"value"`
}

// Validate the Filter.
func (f *Filter) Validate() error {
	if f == nil {
		return nil
	}

	if err := f.Prop.Validate(); err != nil {
		return fmt.Errorf("prop: %s", err)
	}

	if err := f.Prop.Validate(); err != nil {
		return fmt.Errorf("value: %s", err)
	}

	return nil
}

// SearchReq translates the properties as expected by the Search method.
type SearchReq struct {
	Query       string `json:"query,omitempty"`
	StartCursor string `json:"start_cursor,omitempty"`
	PageSize    int32  `json:"page_size,omitempty"`

	Sort   *Sort   `json:"sort,omitempty"`
	Filter *Filter `json:"filter,omitempty"`
}

// Validate the SearchReq.
func (sr *SearchReq) Validate() error {
	switch {
	case sr.PageSize > maxSearchPageSize:
		return fmt.Errorf("page size is %d, max allowed - %d", sr.PageSize, maxSearchPageSize)
	}

	if err := sr.Sort.Validate(); err != nil {
		return err
	}

	if err := sr.Filter.Validate(); err != nil {
		return err
	}

	return nil
}

// SearchRes represents the returned result for the Search action.
type SearchRes struct {
	HasMore    bool    `json:"has_more"`
	NextCursor *string `json:"next_cursor,omitempty"`
	Object     string  `json:"object"`
	Result     []Item  `json:"results,omitempty"`
}

// Item from the SearchRes.
type Item struct {
	Object string `json:"object"`

	*Database
	*Page
}

// UnmarshalJSON implementation.
func (i *Item) UnmarshalJSON(b []byte) error {
	a := struct {
		Object string `json:"object"`
	}{}
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}

	i.Object = a.Object

	switch a.Object {
	case "database":
		i.Database = &Database{}
		return json.Unmarshal(b, i.Database)
	case "page":
		i.Page = &Page{}
		return json.Unmarshal(b, i.Page)
	}
	return fmt.Errorf("unknown item type: %s", a.Object)
}

// MarshalJSON implementation.
func (i *Item) MarshalJSON() ([]byte, error) {
	switch i.Object {
	case "database":
		return json.Marshal(i.Database)
	case "page":
		return json.Marshal(i.Page)
	}

	return nil, fmt.Errorf("unknown item type: %s", i.Object)
}

// Search the content that is available to the token.
func (n *Notion) Search(ctx context.Context, sReq SearchReq) (SearchRes, error) {
	if err := sReq.Validate(); err != nil {
		return SearchRes{}, err
	}
	srJSON, err := json.Marshal(&sReq)
	if err != nil {
		return SearchRes{}, err
	}

	req, err := http.NewRequest(http.MethodPost, defaultAPIPath+"/v1/search", bytes.NewBuffer(srJSON))
	if err != nil {
		return SearchRes{}, err
	}
	req.Header.Set("Notion-Version", defaultNotionAPIVersion)
	req.Header.Set("Authorization", "Bearer "+n.bearer)
	req.Header.Set("Content-Type", "application/json")

	res, err := n.http.Do(req.WithContext(ctx))
	if err != nil {
		return SearchRes{}, err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case 200:
	case 401:
		return SearchRes{}, autocounter.ErrUnauthorized
	default:
		e, _ := io.ReadAll(res.Body)
		return SearchRes{}, fmt.Errorf("unexpected error %d: %s", res.StatusCode, e)
	}

	var sRes SearchRes
	if err := json.NewDecoder(res.Body).Decode(&sRes); err != nil {
		return SearchRes{}, err
	}

	return sRes, nil
}
