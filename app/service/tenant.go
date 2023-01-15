package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	autocounter "github.com/notionplusid/core/app"
	"github.com/notionplusid/core/app/provider/notion"
	"github.com/notionplusid/core/app/storage"
)

// ProcWsFunc is the expected handler for the workspace processing.
type ProcWsFunc func(ctx context.Context, ws autocounter.Workspace) (autocounter.Workspace, error)

// Tenant service.
type Tenant struct {
	s  storage.Storage
	nc notion.ExtConfig
}

// NewTenant constructor.
func NewTenant(s storage.Storage, nc notion.ExtConfig) (*Tenant, error) {
	if err := nc.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	if s == nil {
		return nil, errors.New("storage is required")
	}

	return &Tenant{
		s:  s,
		nc: nc,
	}, nil
}

// Workspace returns the configuration for the provided tenant ID.
func (t *Tenant) Workspace(ctx context.Context, tenantID string) (autocounter.Workspace, error) {
	return t.s.Workspace(ctx, tenantID)
}

// AuthWorkspace by the provided code from the Notion redirect.
func (t *Tenant) AuthWorkspace(ctx context.Context, code string) (autocounter.Workspace, error) {
	res, err := notion.OAuth2(ctx, code, t.nc)
	if err != nil {
		return autocounter.Workspace{}, err
	}

	return autocounter.NewWorkspace(res.WorkspaceID, res.AccessToken)
}

// RegisterWorkspace and persist it to the database.
func (t *Tenant) RegisterWorkspace(ctx context.Context, ws autocounter.Workspace) (autocounter.Workspace, error) {
	return t.s.StoreWorkspace(ctx, ws)
}

// IsAvailable returns error if there's a way to reach out the workspace.
func (t *Tenant) IsAvailable(ctx context.Context, ws autocounter.Workspace) error {
	n, err := notion.NewFromWorkspace(ws)
	if err != nil {
		return err
	}
	defer n.Close()

	_, err = n.Me(ctx)
	return err
}

// UnregisterWorkspace returns nil if deregistration was successful.
func (t *Tenant) UnregisterWorkspace(ctx context.Context, wsID string) error {
	if err := t.s.RemoveWorkspace(ctx, wsID); err != nil {
		return err
	}
	return t.s.RemoveTablesFromWS(ctx, wsID)
}

// ProcOldestUpdated consistently processes the tenants that were processed the longest ago.
func (t *Tenant) ProcOldestUpdated(ctx context.Context, count int64, procWs ProcWsFunc) error {
	return t.s.ProcOldestUpdatedWss(ctx, count, func(ctx context.Context, wss ...autocounter.Workspace) error {
		wg := &sync.WaitGroup{}
		for _, ws := range wss {
			wg.Add(1)
			go func(ws autocounter.Workspace) {
				defer wg.Done()

				err := t.IsAvailable(ctx, ws)
				switch {
				case err == autocounter.ErrUnauthorized:
					if err = t.UnregisterWorkspace(ctx, ws.ID); err != nil {
						log.Printf("Tenant: ProcOldestUpdated: ProcOldestUpdatedWss: couldn't unregister the workspace %s: %s", ws.ID, err)
					}
				case err != nil:
					log.Printf("Tenant: ProcOldestUpdated: ProcOldestUpdatedWss: couldn't check the availability of the workspace %s: %s", ws.ID, err)
				}

				_, err = procWs(ctx, ws)
				if err != nil {
					log.Printf("Tenant: ProcOldestUpdated: ProcOldestUpdatedWss: couldn't process workspace %s: %s", ws.ID, err)
				}
			}(ws)
		}
		wg.Wait()
		return nil
	})
}
