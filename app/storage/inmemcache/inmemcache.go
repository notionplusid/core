package inmemcache

// import (
// 	"context"
// 	"errors"
// 	"sync"
// 	"time"

// 	autocounter "github.com/notionplusid/core/app"
// 	"github.com/notionplusid/core/app/storage"
// )

// const defaultCacheSyncTimeout = 1 * time.Minute

// type cache struct {
// 	wss        []autocounter.Workspace
// 	ts         []autocounter.Table
// 	lastSynced time.Time
// 	mu         *sync.RWMutex
// }

// // Instance wraps a specific storage implementation
// // and instead of writing to storage every time
// // it keeps the value in memory first and writes to storage once in a while.
// type Instance struct {
// 	s storage.Storage
// 	c cache
// }

// var _ storage.Storage = (*Instance)(nil)

// // New Instance constructor.
// func New(s storage.Storage) (*Instance, error) {
// 	if s == nil {
// 		return nil, errors.New("storage instance wasn't provided: nothing to wrap")
// 	}

// 	i := &Instance{
// 		s: s,
// 		c: cache{
// 			mu: &sync.RWMutex{},
// 		},
// 	}

// 	return i, nil
// }

// // Sync all the Workspaces.
// func (i *Instance) Sync(ctx context.Context) error {
// 	i.c.mu.Lock()
// 	defer i.c.mu.Unlock()

// 	wss, err := i.s.Workspaces(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	i.c.wss = wss
// 	return nil
// }

// // Workspace returns cache from the memory.
// func (i *Instance) Workspace(ctx context.Context, id string) (autocounter.Workspace, error) {
// 	if id == "" {
// 		return autocounter.Workspace{}, errors.New("id is required")
// 	}

// 	i.c.mu.RLock()
// 	defer i.c.mu.RUnlock()

// 	for _, item := range i.c.wss {
// 		if item.ID == id {
// 			return item, nil
// 		}
// 	}

// 	return autocounter.Workspace{}, autocounter.ErrNoResults
// }

// // Workspaces values from the cache.
// func (i *Instance) Workspaces(ctx context.Context) ([]autocounter.Workspace, error) {
// 	i.c.mu.RLock()
// 	defer i.c.mu.RUnlock()

// 	return append([]autocounter.Workspace{}, i.c.wss...), nil
// }

// // StoreWorkspace writes the value to the storage first and then also duplicates it into the memory cache.
// func (i *Instance) StoreWorkspace(ctx context.Context, ws autocounter.Workspace) (autocounter.Workspace, error) {
// 	ws, err := i.s.StoreWorkspace(ctx, ws)
// 	if err != nil {
// 		return autocounter.Workspace{}, err
// 	}

// 	i.c.mu.Lock()
// 	defer i.c.mu.Unlock()

// 	// find and override the workspace.
// 	for n, item := range i.c.wss {
// 		if item.ID == ws.ID {
// 			i.c.wss[n] = ws
// 			return ws, nil
// 		}
// 	}
// 	// if doesn't exist yet - save it to cache.
// 	i.c.wss = append(i.c.wss, ws)

// 	return ws, nil
// }

// // ProcOldestUpdatedWss works mostly with in-memory values
// // and once a minute processes the values from the database
// // while updating the cache.
// func (i *Instance) ProcOldestUpdatedWss(ctx context.Context, count int64, procWss storage.ProcWssFunc) error {

// }

// func (i *Instance) RemoveWorkspace(ctx context.Context, wsID string) error {

// }

// func (i *Instance) StoreTable(ctx context.Context, workspaceID string, table autocounter.Table) (autocounter.Table, error) {

// }

// func (i *Instance) DisableTable(ctx context.Context, wsID, tableID string) error {

// }

// func (i *Instance) ActiveTables(ctx context.Context, workspaceID string, tableIDs []string) ([]string, error) {

// }

// func (i *Instance) ListAllActiveTables(ctx context.Context, workspaceID string) ([]autocounter.Table, error) {

// }

// func (i *Instance) RemoveTablesFromWS(ctx context.Context, wsID string) error {

// }
