package inmemcache

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	autocounter "github.com/notionplusid/core/app"
	"github.com/notionplusid/core/app/storage"
)

const defaultCacheSyncTimeout = 5 * time.Minute

type cache struct {
	wss []autocounter.Workspace
	ts  []autocounter.Table
	mu  *sync.RWMutex
}

// Instance wraps a specific storage implementation
// and instead of writing to storage every time
// it keeps the value in memory first and writes to storage once in a while.
type Instance struct {
	s storage.Storage
	c cache
}

var _ storage.Storage = (*Instance)(nil)

// New Instance constructor.
func New(s storage.Storage) (*Instance, error) {
	if s == nil {
		return nil, errors.New("storage instance wasn't provided: nothing to wrap")
	}

	i := &Instance{
		s: s,
		c: cache{
			mu: &sync.RWMutex{},
		},
	}

	return i, nil
}

// Sync all the Workspaces.
func (i *Instance) Sync(ctx context.Context) error {
	i.c.mu.Lock()
	defer i.c.mu.Unlock()

	wss, err := i.s.Workspaces(ctx)
	if err != nil {
		return err
	}
	i.c.wss = wss

	ts, err := i.s.Tables(ctx)
	if err != nil {
		return err
	}
	i.c.ts = ts

	return nil
}

// Workspace returns cache from the memory.
func (i *Instance) Workspace(ctx context.Context, id string) (autocounter.Workspace, error) {
	if id == "" {
		return autocounter.Workspace{}, errors.New("id is required")
	}

	i.c.mu.RLock()
	defer i.c.mu.RUnlock()

	for _, item := range i.c.wss {
		if item.ID == id {
			return item, nil
		}
	}

	return autocounter.Workspace{}, autocounter.ErrNoResults
}

// Workspaces values from the cache.
func (i *Instance) Workspaces(ctx context.Context) ([]autocounter.Workspace, error) {
	i.c.mu.RLock()
	defer i.c.mu.RUnlock()

	return append([]autocounter.Workspace{}, i.c.wss...), nil
}

// Tables values from the cache.
func (i *Instance) Tables(ctx context.Context) ([]autocounter.Table, error) {
	i.c.mu.RLock()
	defer i.c.mu.RUnlock()

	return append([]autocounter.Table{}, i.c.ts...), nil
}

// StoreWorkspace writes the value to the storage first and then also duplicates it into the memory cache.
func (i *Instance) StoreWorkspace(ctx context.Context, ws autocounter.Workspace) (autocounter.Workspace, error) {
	ws, err := i.s.StoreWorkspace(ctx, ws)
	if err != nil {
		return autocounter.Workspace{}, err
	}

	i.c.mu.Lock()
	defer i.c.mu.Unlock()

	// find and override the workspace.
	for n, item := range i.c.wss {
		if item.ID == ws.ID {
			i.c.wss[n] = ws
			return ws, nil
		}
	}
	// if doesn't exist yet - save it to cache.
	i.c.wss = append(i.c.wss, ws)

	return ws, nil
}

// ProcOldestUpdatedWss works mostly with in-memory values
// and once a minute processes the values from the database
// while updating the cache.
func (i *Instance) ProcOldestUpdatedWss(ctx context.Context, count int64, procWss storage.ProcWssFunc) error {
	// if cache has some old items instances - we run proper sync until we have cache in the most recent state.
	if i.c.oldestProcessedWs().Add(defaultCacheSyncTimeout).Before(time.Now()) {
		defer i.Sync(ctx) // nolint: errcheck
		return i.s.ProcOldestUpdatedWss(ctx, count, func(ctx context.Context, wss ...autocounter.Workspace) error {
			defer func() {
				for _, ws := range wss {
					if err := i.c.updateWs(ws); err != nil {
						log.Printf("In-mem cache: couldn't update workspace %s: %s", ws.ID, err)
					}
				}
			}()
			return procWss(ctx, wss...)
		})
	}

	i.c.mu.RLock()
	wss := append([]autocounter.Workspace{}, i.c.wss...)
	i.c.mu.RUnlock()

	// if cache is relatively recent - make the call purely from cache.
	for _, ws := range wss {
		if err := i.c.updateWs(ws); err != nil {
			log.Printf("In-mem cache: couldn't update workspace %s: %s", ws.ID, err)
		}
	}
	return nil
}

func (c *cache) updateWs(ws autocounter.Workspace) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, item := range c.wss {
		if item.ID != ws.ID {
			continue
		}

		ws.ProcessedAt = time.Now()
		c.wss[i] = ws
		return nil
	}

	return fmt.Errorf("no cached workspace with id %s", ws.ID)
}

func (c *cache) updateTable(t autocounter.Table) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, item := range c.ts {
		if item.ID != t.ID && item.WorkspaceID != t.WorkspaceID {
			continue
		}

		c.ts[i] = t
		return nil
	}

	return fmt.Errorf("no cached table with id %s", t.ID)
}

func (c *cache) oldestProcessedWs() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.wss) == 0 {
		return time.Time{}
	}

	date := c.wss[0].ProcessedAt
	for _, ws := range c.wss {
		if ws.ProcessedAt.After(date) {
			continue
		}

		date = ws.ProcessedAt
	}

	return date
}

func (c *cache) removeWs(wsID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, ws := range c.wss {
		if ws.ID != wsID {
			continue
		}

		c.wss = append(c.wss[:i], c.wss[i+1:]...)
		return
	}
}

// RemoveWorkspace from cache and the database.
func (i *Instance) RemoveWorkspace(ctx context.Context, wsID string) error {
	defer i.c.removeWs(wsID)
	return i.s.RemoveWorkspace(ctx, wsID)
}

// StoreTable into the cache and the database.
func (i *Instance) StoreTable(ctx context.Context, workspaceID string, table autocounter.Table) (autocounter.Table, error) {
	table, err := i.s.StoreTable(ctx, workspaceID, table)
	if err != nil {
		return autocounter.Table{}, err
	}

	i.c.mu.Lock()
	defer i.c.mu.Unlock()

	// find and override the workspace.
	for n, item := range i.c.ts {
		if item.ID == table.ID && item.WorkspaceID == table.WorkspaceID {
			i.c.ts[n] = table
			return table, nil
		}
	}
	// if doesn't exist yet - save it to cache.
	i.c.ts = append(i.c.ts, table)

	return table, nil
}

func (i *Instance) DisableTable(ctx context.Context, wsID, tableID string) (autocounter.Table, error) {
	t, err := i.s.DisableTable(ctx, wsID, tableID)
	if err != nil {
		return autocounter.Table{}, err
	}

	return t, i.c.updateTable(t)
}

func (i *Instance) ActiveTables(ctx context.Context, workspaceID string, tableIDs []string) ([]string, error) {
	ids := map[string]struct{}{}
	for _, id := range tableIDs {
		ids[id] = struct{}{}
	}

	i.c.mu.RLock()
	defer i.c.mu.RUnlock()

	var activeIDs []string
	for _, t := range i.c.ts {
		// in case if such ID is registered with another workspace.
		if t.WorkspaceID != workspaceID {
			continue
		}

		if t.Status != autocounter.StatusActive {
			continue
		}

		_, ok := ids[t.ID]
		if !ok {
			continue
		}

		activeIDs = append(activeIDs, t.ID)
	}

	return activeIDs, nil
}

func (i *Instance) ListAllActiveTables(ctx context.Context, workspaceID string) ([]autocounter.Table, error) {
	i.c.mu.RLock()
	defer i.c.mu.RUnlock()

	var activeTables []autocounter.Table
	for _, t := range i.c.ts {
		// in case if such ID is registered with another workspace.
		if t.WorkspaceID != workspaceID {
			continue
		}

		if t.Status != autocounter.StatusActive {
			continue
		}

		tl := t
		activeTables = append(activeTables, tl)
	}

	return activeTables, nil
}

func (i *Instance) RemoveTablesFromWS(ctx context.Context, wsID string) error {
	err := i.s.RemoveTablesFromWS(ctx, wsID)
	if err != nil {
		return err
	}

	i.c.mu.Lock()
	defer i.c.mu.Unlock()

	var ts []autocounter.Table
	for _, t := range i.c.ts {
		if t.WorkspaceID == wsID {
			continue
		}

		tl := t
		ts = append(ts, tl)
	}
	i.c.ts = ts
	return nil
}
