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

	autocounter "github.com/notionplusid/core/main"
)

const maxDBQueryPageSize = 100

// PropertyType within the Database.
type PropertyType string

// Known PropertyTypes.
const (
	PropertyTypeTitle          = "title"
	PropertyTypeRichText       = "rich_text"
	PropertyTypeNumber         = "number"
	PropertyTypeSelect         = "select"
	PropertyTypeMultiSelect    = "multi_select"
	PropertyTypeDate           = "date"
	PropertyTypePeople         = "people"
	PropertyTypeFile           = "file"
	PropertyTypeCheckbox       = "checkbox"
	PropertyTypeURL            = "url"
	PropertyTypeEmail          = "email"
	PropertyTypePhoneNumber    = "phone_number"
	PropertyTypeFormula        = "formula"
	PropertyTypeRelation       = "relation"
	PropertyTypeRollup         = "rollup"
	PropertyTypeCreatedTime    = "created_time"
	PropertyTypeCreatedBy      = "created_by"
	PropertyTypeLastEditedTime = "last_edited_time"
	PropertyTypeLastEditedBy   = "last_edited_by"
)

// Database object.
type Database struct {
	ID             string     `json:"id"`
	Object         string     `json:"object"`
	CreatedTime    time.Time  `json:"created_time"`
	LastEditedTime time.Time  `json:"last_edited_at"`
	Title          []RichText `json:"title"`
	Properties     map[string]struct {
		ID     string       `json:"id"`
		Type   PropertyType `json:"type"`
		Number *struct {
			Format string `json:"format"`
		} `json:"number,omitempty"`
		Formula *struct {
			Expression string `json:"expression,omitempty"`
		} `json:"formula,omitempty"`
		Relation *struct {
			DatabaseID         string `json:"database_id"`
			SyncedPropertyName string `json:"synced_property_name,omitempty"`
			SyncedPropertyID   string `json:"synced_property_id,omitempty"`
		} `json:"relation,omitempty"`
		Rollup *struct {
			RelationPropertyName string `json:"relation_property_name"`
			RelationPropertyID   string `json:"relation_property_id"`
			RollupPropertyName   string `json:"rollup_property_name"`
			RollupPropertyID     string `json:"rollup_property_id"`
			Function             string `json:"function"`
		} `json:"rollup,omitempty"`
		Select *struct {
			Options []struct {
				ID    string `json:"id"`
				Name  string `json:"name"`
				Color string `json:"color"`
			} `json:"options,omitempty"`
		} `json:"select,omitempty"`
		MultiSelect *struct {
			Options []struct {
				ID    string `json:"id"`
				Name  string `json:"name"`
				Color string `json:"color"`
			} `json:"options,omitempty"`
		} `json:"multi_select,omitempty"`
	} `json:"properties"`
	Parent struct {
		Type      string `json:"type"`
		PageID    string `json:"page_id,omitempty"`
		Workspace bool   `json:"workspace,omitempty"`
	} `json:"parent"`
}

type DBSortTimestamp string

const (
	DBSortTimestampCreated    DBSortTimestamp = "created_time"
	DBSortTimestampLastEdited DBSortTimestamp = "last_edited_time"
)

type DBSortDirection string

const (
	DBSortDirectionAsc  DBSortDirection = "ascending"
	DBSortDirectionDesc DBSortDirection = "descending"
)

type DBSort struct {
	Property  string          `json:"property,omitempty"`
	Timestamp DBSortTimestamp `json:"timestamp,omitempty"`
	Direction DBSortDirection `json:"direction"`
}

type DBFilterText struct {
	Equals         *string `json:"equals,omitempty"`
	DoesNotEqual   *string `json:"does_not_equal,omitempty"`
	Contains       string  `json:"contains,omitempty"`
	DoesNotContain string  `json:"does_not_contain,omitempty"`
	StartsWith     string  `json:"starts_with,omitempty"`
	EndsWith       string  `json:"ends_with,omitempty"`
	IsEmpty        bool    `json:"is_empty,omitempty"`
	IsNotEmpty     bool    `json:"is_not_empty,omitempty"`
}

type DBFilterNumber struct {
	Equals               *float64 `json:"equals,omitempty"`
	DoesNotEqual         *float64 `json:"does_not_equal,omitempty"`
	GreaterThan          *float64 `json:"greater_than,omitempty"`
	LessThan             *float64 `json:"less_than,omitempty"`
	GreaterThanOrEqualTo *float64 `json:"greater_than_or_equal_to,omitempty"`
	LessThanOrEqualTo    *float64 `json:"less_than_or_equal_to,omitempty"`
	IsEmpty              bool     `json:"is_empty,omitempty"`
	IsNotEmpty           bool     `json:"is_not_empty,omitempty"`
}

type DBFilterCheckbox struct {
	Equals       *bool `json:"equals,omitempty"`
	DoesNotEqual *bool `json:"does_not_equal,omitempty"`
}

type DBFilterSelect struct {
	Equals       *string `json:"equals,omitempty"`
	DoesNotEqual *string `json:"does_not_equal,omitempty"`
	IsEmpty      bool    `json:"is_empty"`
	IsNotEmpty   bool    `json:"is_not_empty"`
}

type DBFilterMultiSelect struct {
	Equals       *string `json:"equals,omitempty"`
	DoesNotEqual *string `json:"does_not_equal,omitempty"`
	IsEmpty      bool    `json:"is_empty"`
	IsNotEmpty   bool    `json:"is_not_empty"`
}

type DBFilterDate struct {
	Equals     *time.Time `json:"equals,omitempty"`
	Before     *time.Time `json:"before,omitempty"`
	After      *time.Time `json:"after,omitempty"`
	OnOrBefore *time.Time `json:"on_or_before,omitempty"`
	OnOrAfter  *time.Time `json:"on_or_after,omitempty"`
	IsEmpty    bool       `json:"is_empty,omitempty"`
	IsNotEmpty bool       `json:"is_not_empty,omitempty"`
	PastWeek   *struct{}  `json:"past_week,omitempty"`
	PastMonth  *struct{}  `json:"past_month,omitempty"`
	PastYear   *struct{}  `json:"past_year,omitempty"`
	NextWeek   *struct{}  `json:"next_week,omitempty"`
	NextMonth  *struct{}  `json:"next_month,omitempty"`
	NextYear   *struct{}  `json:"next_year,omitempty"`
}

type DBFilterPeople struct {
	Contains       string `json:"contains,omitempty"`
	DoesNotContain string `json:"does_not_contain,omitempty"`
	IsEmpty        bool   `json:"is_empty,omitempty"`
	IsNotEmpty     bool   `json:"is_not_empty,omitempty"`
}

type DBFilterFiles struct {
	IsEmpty    bool `json:"is_empty,omitempty"`
	IsNotEmpty bool `json:"is_not_empty,omitempty"`
}

type DBFilterRelation struct {
	Contains       string `json:"contains,omitempty"`
	DoesNotContain string `json:"does_not_contain,omitempty"`
	IsEmpty        bool   `json:"is_empty,omitempty"`
	IsNotEmpty     bool   `json:"is_not_empty,omitempty"`
}

type DBFilterFormula struct {
	Text     *DBFilterText     `json:"text,omitempty"`
	Checkbox *DBFilterCheckbox `json:"checkbox,omitempty"`
	Number   *DBFilterNumber   `json:"number,omitempty"`
	Date     *DBFilterDate     `json:"date,omitempty"`
}

type DBFilter struct {
	Property       string               `json:"property"`
	Title          *DBFilterText        `json:"title,omitempty"`
	RichText       *DBFilterText        `json:"rich_text,omitempty"`
	URL            *DBFilterText        `json:"url,omitempty"`
	Email          *DBFilterText        `json:"email,omitempty"`
	Phone          *DBFilterText        `json:"phone,omitempty"`
	Number         *DBFilterNumber      `json:"number,omitempty"`
	Checkbox       *DBFilterCheckbox    `json:"checkbox,omitempty"`
	Select         *DBFilterSelect      `json:"select,omitempty"`
	MultiSelect    *DBFilterMultiSelect `json:"multi_select,omitempty"`
	Date           *DBFilterDate        `json:"date,omitempty"`
	CreatedTime    *DBFilterDate        `json:"created_time,omitempty"`
	LastEditedTime *DBFilterDate        `json:"last_edited_time,omitempty"`
	People         *DBFilterPeople      `json:"people,omitempty"`
	CreatedBy      *DBFilterPeople      `json:"created_by,omitempty"`
	LastEditedBy   *DBFilterPeople      `json:"last_edited_by,omitempty"`
	Files          *DBFilterFiles       `json:"files,omitempty"`
	Relation       *DBFilterRelation    `json:"relation,omitempty"`
	Formula        *DBFilterFormula     `json:"formula,omitempty"`

	Or  []DBFilter `json:"or,omitempty"`
	And []DBFilter `json:"and,omitempty"`
}

// DBQueryReq is a database query request sent to the API.
type DBQueryReq struct {
	Filter      *DBFilter `json:"filter,omitempty"`
	Sorts       []DBSort  `json:"sorts,omitempty"`
	StartCursor string    `json:"start_cursor,omitempty"`
	PageSize    int32     `json:"page_size,omitempty"`
}

// Validate the DBQueryReq.
func (req *DBQueryReq) Validate() error {
	if req.PageSize > maxDBQueryPageSize {
		return fmt.Errorf("page size is too big: max - %d: provided - %d", maxDBQueryPageSize, req.PageSize)
	}

	return nil
}

// DBQueryReq is a database query response received from the API.
type DBQueryRes struct {
	HasMore    bool    `json:"has_more"`
	NextCursor *string `json:"next_cursor,omitempty"`
	Object     string  `json:"object"`
	Result     []Page  `json:"results,omitempty"`
}

// QueryDatabase returns result Query result that matches the requirements from DBQueryReq.
func (n *Notion) QueryDatabase(ctx context.Context, databaseID string, q DBQueryReq) (DBQueryRes, error) {
	if databaseID == "" {
		return DBQueryRes{}, errors.New("database id is required")
	}
	if err := q.Validate(); err != nil {
		return DBQueryRes{}, err
	}
	srJSON, err := json.Marshal(&q)
	if err != nil {
		return DBQueryRes{}, err
	}

	req, err := http.NewRequest(http.MethodPost, defaultAPIPath+"/v1/databases/"+databaseID+"/query", bytes.NewBuffer(srJSON))
	if err != nil {
		return DBQueryRes{}, err
	}
	req.Header.Set("Notion-Version", defaultNotionAPIVersion)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+n.bearer)

	res, err := n.http.Do(req.WithContext(ctx))
	if err != nil {
		return DBQueryRes{}, err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case 200:
	case 400:
		return DBQueryRes{}, autocounter.ErrIncompatibleTable
	case 404:
		return DBQueryRes{}, autocounter.ErrTableNotFound
	default:
		e, _ := io.ReadAll(res.Body)
		return DBQueryRes{}, fmt.Errorf("unexpected error %d: %s", res.StatusCode, e)
	}

	var qRes DBQueryRes
	if err := json.NewDecoder(res.Body).Decode(&qRes); err != nil {
		return DBQueryRes{}, err
	}

	return qRes, nil
}

// Database returns the resource of a database from the Notion API.
func (n *Notion) Database(ctx context.Context, databaseID string) (Database, error) {
	if databaseID == "" {
		return Database{}, errors.New("database id is required")
	}

	req, err := http.NewRequest(http.MethodGet, defaultAPIPath+"/v1/databases/"+databaseID, nil)
	if err != nil {
		return Database{}, err
	}
	req.Header.Set("Notion-Version", defaultNotionAPIVersion)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+n.bearer)

	res, err := n.http.Do(req.WithContext(ctx))
	if err != nil {
		return Database{}, err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case 200:
	case 404:
		return Database{}, autocounter.ErrTableNotFound
	default:
		e, _ := io.ReadAll(res.Body)
		return Database{}, fmt.Errorf("unexpected error %d: %s", res.StatusCode, e)
	}

	var qRes Database
	if err := json.NewDecoder(res.Body).Decode(&qRes); err != nil {
		return Database{}, err
	}

	return qRes, nil
}
