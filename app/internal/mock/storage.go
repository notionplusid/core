package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	autocounter "github.com/notionplusid/core/app"
	"github.com/notionplusid/core/app/storage"
)

var _ storage.Storage = (*Storage)(nil)

type Storage struct {
	mock.Mock
}

func (s *Storage) Workspace(ctx context.Context, id string) (autocounter.Workspace, error) {
	args := s.Called(ctx, id)
	return args.Get(0).(autocounter.Workspace), args.Error(1)
}

func (s *Storage) Workspaces(ctx context.Context) ([]autocounter.Workspace, error) {
	args := s.Called(ctx)
	return args.Get(0).([]autocounter.Workspace), args.Error(1)
}

func (s *Storage) StoreWorkspace(ctx context.Context, ws autocounter.Workspace) (autocounter.Workspace, error) {
	args := s.Called(ctx, ws)
	return args.Get(0).(autocounter.Workspace), args.Error(1)
}

func (s *Storage) ProcOldestUpdatedWss(ctx context.Context, count int64, procWss storage.ProcWssFunc) error {
	args := s.Called(ctx, count, procWss)
	return args.Error(0)
}

func (s *Storage) RemoveWorkspace(ctx context.Context, wsID string) error {
	args := s.Called(ctx, wsID)
	return args.Error(0)
}

func (s *Storage) Tables(ctx context.Context) ([]autocounter.Table, error) {
	args := s.Called(ctx)
	return args.Get(0).([]autocounter.Table), args.Error(1)
}

func (s *Storage) StoreTable(ctx context.Context, workspaceID string, table autocounter.Table) (autocounter.Table, error) {
	args := s.Called(ctx, workspaceID, table)
	return args.Get(0).(autocounter.Table), args.Error(1)
}

func (s *Storage) DisableTable(ctx context.Context, wsID, tableID string) (autocounter.Table, error) {
	args := s.Called(ctx, wsID, tableID)
	return args.Get(0).(autocounter.Table), args.Error(1)
}

func (s *Storage) ActiveTables(ctx context.Context, workspaceID string, tableIDs []string) ([]string, error) {
	args := s.Called(ctx, workspaceID, tableIDs)
	return args.Get(0).([]string), args.Error(1)
}

func (s *Storage) ListAllActiveTables(ctx context.Context, workspaceID string) ([]autocounter.Table, error) {
	args := s.Called(ctx, workspaceID)
	return args.Get(0).([]autocounter.Table), args.Error(1)
}

func (s *Storage) RemoveTablesFromWS(ctx context.Context, wsID string) error {
	args := s.Called(ctx, wsID)
	return args.Error(0)
}
