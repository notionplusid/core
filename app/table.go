package autocounter

import (
	"errors"
	"fmt"
	"time"
)

const DefaultTableParamName = "PlusID"

// Status of the Table observation.
type Status = string

// Known statuses.
const (
	StatusActive   Status = "active"
	StatusDisabled Status = "disabled"
)

var validStatuses = []Status{
	StatusActive,
	StatusDisabled,
}

// ValidateStatus and return error if provided status is invalid.
func ValidateStatus(s Status) error {
	if s == "" {
		return errors.New("status is required")
	}

	for _, vs := range validStatuses {
		if s == vs {
			return nil
		}
	}

	return fmt.Errorf("unknown status: %s", s)
}

// Table domain structure.
type Table struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspaceId"`
	Status      Status    `json:"status"`
	ParamName   string    `json:"paramName,omitempty"`
	CreatedAt   time.Time `json:"createdAt,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt,omitempty"`
}

// New Table constructor.
func New(id, workspaceID string) (Table, error) {
	switch {
	case id == "":
		return Table{}, errors.New("id is required")
	case workspaceID == "":
		return Table{}, errors.New("workspace is required")
	}

	return Table{
		ID:          id,
		WorkspaceID: workspaceID,
		Status:      StatusActive,
		ParamName:   DefaultTableParamName,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// Validate the Table.
func (t *Table) Validate() error {
	switch {
	case t.ID == "":
		return errors.New("id is required")
	case t.WorkspaceID == "":
		return errors.New("workspace id is required")
	case t.Status == "":
		return errors.New("status id is required")
	case t.ParamName == "":
		return errors.New("param name is required")
	}

	return nil
}

// Diff returns elements that are present in the set and are not present in the subset.
func Diff(set, subset []string) []string {
	if len(set) == 0 {
		return []string{}
	}

	if len(subset) == 0 {
		return set
	}

	diffMap := map[string]struct{}{}
	for _, s := range subset {
		if _, ok := diffMap[s]; ok {
			continue
		}

		diffMap[s] = struct{}{}
	}

	var diff []string
	for _, s := range set {
		if _, ok := diffMap[s]; ok {
			continue
		}

		diff = append(diff, s)
	}

	return diff
}
