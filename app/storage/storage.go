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
	RemoveWorkspace(ctx context.Context, wsID string) error

	Tables(ctx context.Context) ([]autocounter.Table, error)
	StoreTable(ctx context.Context, workspaceID string, table autocounter.Table) (autocounter.Table, error)
	DisableTable(ctx context.Context, wsID, tableID string) (autocounter.Table, error)
	ActiveTables(ctx context.Context, workspaceID string, tableIDs []string) ([]string, error)
	ListAllActiveTables(ctx context.Context, workspaceID string) ([]autocounter.Table, error)
	RemoveTablesFromWS(ctx context.Context, wsID string) error
}
