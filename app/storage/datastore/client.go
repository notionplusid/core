package datastore

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	datastoresdk "cloud.google.com/go/datastore"

	autocounter "github.com/notionplusid/core/app"
	"github.com/notionplusid/core/app/storage"
)

const (
	workspaceKey = "Workspace"
	tableKey     = "Table"
)

// Client for Datastore.
type Client struct {
	ds *datastoresdk.Client
}

// Make sure Datastore client is compatible.
var _ storage.Storage = (*Client)(nil)

// New datastore Client constructor.
func New(ctx context.Context, projectName string) (*Client, error) {
	c, err := datastoresdk.NewClient(ctx, projectName)
	if err != nil {
		return nil, err
	}

	return &Client{ds: c}, nil
}

// Workspace returns the instance by the requested ID.
func (c *Client) Workspace(ctx context.Context, id string) (autocounter.Workspace, error) {
	if id == "" {
		return autocounter.Workspace{}, errors.New("id is required")
	}

	var res autocounter.Workspace
	err := c.ds.Get(ctx, datastoresdk.NameKey(workspaceKey, id, nil), &res)
	switch {
	case err == datastoresdk.ErrNoSuchEntity:
		return autocounter.Workspace{}, autocounter.ErrNoResults
	case err != nil:
		return autocounter.Workspace{}, err
	}

	return res, nil
}

// StoreWorkspace for future usage along with the access token.
func (c *Client) StoreWorkspace(ctx context.Context, ws autocounter.Workspace) (autocounter.Workspace, error) {
	if err := ws.Validate(); err != nil {
		return autocounter.Workspace{}, err
	}

	now := time.Now()
	ws.CreatedAt = now
	ws.UpdatedAt = now

	_, err := c.ds.Put(ctx, datastoresdk.NameKey(workspaceKey, ws.ID, nil), &ws)
	if err != nil {
		return autocounter.Workspace{}, fmt.Errorf("couldn't put the name key: %s", err)
	}

	return ws, nil
}

// ListAllTables stored in the database for the workspace.
func (c *Client) ListAllTables(ctx context.Context, workspaceID string) ([]autocounter.Table, error) {
	if workspaceID == "" {
		return nil, errors.New("workspace id is required")
	}

	key := datastoresdk.
		NewQuery(tableKey).
		FilterField("WorkspaceID", "=", workspaceID)

	var res []autocounter.Table
	_, err := c.ds.GetAll(ctx, key, &res)
	switch {
	case err == datastoresdk.ErrNoSuchEntity:
		return nil, autocounter.ErrNoResults
	case err != nil:
		return nil, err
	}

	return res, nil
}

// Workspaces returns all the available instances.
func (c *Client) Workspaces(ctx context.Context) ([]autocounter.Workspace, error) {
	var res []autocounter.Workspace
	_, err := c.ds.GetAll(ctx, datastoresdk.NewQuery(workspaceKey), &res)
	switch {
	case err == datastoresdk.ErrNoSuchEntity:
		return nil, autocounter.ErrNoResults
	case err != nil:
		return nil, err
	}

	return res, nil
}

// DisableTable instance.
func (c *Client) DisableTable(ctx context.Context, wsID, tID string) error {
	_, err := c.ds.RunInTransaction(ctx, func(tx *datastoresdk.Transaction) error {
		key := datastoresdk.NameKey(tableKey, tID, nil)

		var t autocounter.Table
		err := tx.Get(key, &t)
		switch {
		case err == datastoresdk.ErrNoSuchEntity:
			return autocounter.ErrNoResults
		case err != nil:
			return err
		}

		if t.WorkspaceID != wsID {
			return autocounter.ErrNoResults
		}

		t.Status = autocounter.StatusDisabled

		_, err = tx.Mutate(datastoresdk.NewUpdate(key, &t))
		return err
	})
	return err
}

// StoreTable instance.
func (c *Client) StoreTable(ctx context.Context, workspaceID string, table autocounter.Table) (autocounter.Table, error) {
	table.WorkspaceID = workspaceID
	if err := table.Validate(); err != nil {
		return autocounter.Table{}, err
	}

	now := time.Now()
	table.CreatedAt = now
	table.UpdatedAt = now

	_, err := c.ds.Put(ctx, datastoresdk.NameKey(tableKey, table.ID, nil), &table)
	if err != nil {
		return autocounter.Table{}, err
	}
	return table, nil
}

// RegisteredTables returns the subset of table IDs from the provided list that are registered.
func (c *Client) RegisteredTables(ctx context.Context, workspaceID string, tableIDs []string) ([]string, error) {
	var keys []*datastoresdk.Key
	for _, t := range tableIDs {
		keys = append(keys, datastoresdk.NameKey(tableKey, t, nil))
	}

	res := make([]autocounter.Table, len(keys))
	err := c.ds.GetMulti(ctx, keys, res)
	me, ok := err.(datastoresdk.MultiError)
	if ok {
		for _, err := range me {
			if err == nil {
				continue
			}

			log.Printf("Datastore: RegisteredTables: couldn't fetch item: %s", err)
		}
	}
	switch {
	case ok:
	case err != nil:
		return nil, err
	}
	if len(res) < 1 {
		return nil, autocounter.ErrNoResults
	}

	var ids []string
	for _, t := range res {
		// in case if such ID is registered with another workspace.
		if t.WorkspaceID != workspaceID {
			continue
		}

		ids = append(ids, t.ID)
	}

	return ids, nil
}

// ProcOldestUpdatedWss fetches the amount of workspaces that were processed the longest ago and applies the `procWss`
// at the returned workspaces set in a transactional manner.
func (c *Client) ProcOldestUpdatedWss(ctx context.Context, count int64, procWss storage.ProcWssFunc) error {
	_, err := c.ds.RunInTransaction(ctx, func(tx *datastoresdk.Transaction) error {
		// lock items into transaction.
		q := datastoresdk.NewQuery(workspaceKey).
			Transaction(tx).
			Order("ProcessedAt").
			Limit(int(count))

		// fetch all the items that needs to be updated.
		var wss []autocounter.Workspace
		ks, err := c.ds.GetAll(ctx, q, &wss)
		if err != nil {
			return err
		}

		// process every workspace.
		if err := procWss(ctx, wss...); err != nil {
			return err
		}

		// update the processed timestamp.
		now := time.Now()
		for i := range wss {
			wss[i].ProcessedAt = now
		}

		// write the new state back into the database.
		_, err = tx.PutMulti(ks, wss)
		return err
	})
	return err
}
