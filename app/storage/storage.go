package storage

import (
	"context"

	autocounter "github.com/notionplusid/core/app"
)

// ProcWssFunc is the processing handler of the set of workspaces returned from the database.
type ProcWssFunc func(ctx context.Context, wss ...autocounter.Workspace) error

// Storage layer abstraction.
type Storage interface {
	Workspace(ctx context.Context, id string) (autocounter.Workspace, error)
	Workspaces(ctx context.Context) ([]autocounter.Workspace, error)
	StoreWorkspace(ctx context.Context, ws autocounter.Workspace) (autocounter.Workspace, error)
	ProcOldestUpdatedWss(ctx context.Context, count int64, procWss ProcWssFunc) error

	StoreTable(ctx context.Context, workspaceID string, table autocounter.Table) (autocounter.Table, error)
	DisableTable(ctx context.Context, wsID, tableID string) error
	ActiveTables(ctx context.Context, workspaceID string, tableIDs []string) ([]string, error)
	ListAllActiveTables(ctx context.Context, workspaceID string) ([]autocounter.Table, error)
}
