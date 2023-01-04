package autocounter

import "errors"

// Known errors values.
var (
	ErrTableNotFound     error = errors.New("table not found")
	ErrTableIsEmpty      error = errors.New("table is empty")
	ErrPageNotFound      error = errors.New("page not found")
	ErrInvalidTableParam error = errors.New("invalid table parameter")
	ErrNoResults         error = errors.New("no results")
	ErrIncompatibleTable error = errors.New("incompatbile table")
)
