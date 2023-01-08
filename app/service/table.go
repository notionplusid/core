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

const defaultBatchSize = 100
const defaultUpdateChunkSize = 3

// Table service.
type Table struct {
	s         storage.Storage
	batchSize int64
}

// NewTable service constructor.
func NewTable(s storage.Storage) (*Table, error) {
	return &Table{
		s:         s,
		batchSize: defaultBatchSize,
	}, nil
}

// FetchForWs returns Table that can be found in the provided workspace.
func (t *Table) FetchForWs(ctx context.Context, workspaceID, tableID string) (autocounter.Table, error) {
	return autocounter.Table{
		ID:          tableID,
		WorkspaceID: workspaceID,
		Status:      autocounter.StatusActive,
		ParamName:   autocounter.DefaultTableParamName,
	}, nil
}

// IsFillable returns no error in case if the table is eligable for the autoincrement fill.
func (t *Table) IsFillable(ctx context.Context, tableID string, ws autocounter.Workspace) error {
	if tableID == "" {
		return errors.New("table id is required")
	}

	table, err := t.FetchForWs(ctx, ws.ID, tableID)
	if err != nil {
		return fmt.Errorf("couldn't fetch table: %w", err)
	}

	n, err := notion.NewFromWorkspace(ws)
	if err != nil {
		return fmt.Errorf("couldn't initialize notion api client: %s", err)
	}

	db, err := n.Database(ctx, tableID)
	if err != nil {
		return fmt.Errorf("couldn't fetch table information: %s", err)
	}

	p, ok := db.Properties[table.ParamName]
	if !ok {
		return fmt.Errorf("%w: missing column: %s", autocounter.ErrInvalidTableParam, table.ParamName)
	}
	if p.Type != notion.PropertyTypeNumber {
		return fmt.Errorf("%w: wrong type of column %s: %s", autocounter.ErrInvalidTableParam, table.ParamName, p.Type)
	}

	return validateExpectedDatabaseParam(db, table.ParamName)
}

func validateExpectedDatabaseParam(db notion.Database, paramName string) error {
	p, ok := db.Properties[paramName]
	if !ok {
		return fmt.Errorf("%w: missing column: %s", autocounter.ErrInvalidTableParam, paramName)
	}
	if p.Type != notion.PropertyTypeNumber {
		return fmt.Errorf("%w: wrong type of column %s: %s", autocounter.ErrInvalidTableParam, paramName, p.Type)
	}
	return nil
}

// Active returns the list of table IDs that are registered and active for the provided workspace.
func (t *Table) Active(ctx context.Context, workspaceID string, tableIDs []string) ([]string, error) {
	return t.s.ActiveTables(ctx, workspaceID, tableIDs)
}

// Register the Table within the Workspace for further scans and autocounter fills.
func (t *Table) Register(ctx context.Context, workspaceID string, table autocounter.Table) error {
	_, err := t.s.StoreTable(ctx, workspaceID, table)
	return err
}

// Available databases for read.
func (t *Table) Available(ctx context.Context, ws autocounter.Workspace) ([]autocounter.Table, error) {
	if err := ws.Validate(); err != nil {
		return nil, err
	}

	notionCli, err := notion.NewFromWorkspace(ws)
	if err != nil {
		return nil, fmt.Errorf("couldn't initialize notion api client: %s", err)
	}

	res, err := notionCli.Search(ctx, notion.SearchReq{
		Filter: &notion.Filter{
			Prop:  "object",
			Value: "database",
		},
		Sort: &notion.Sort{
			Direction: notion.SortDirectionDesc,
			Timestamp: notion.SortTimestampLastEdited,
		},
		PageSize: 100,
	})
	if err != nil {
		return nil, err
	}

	if len(res.Result) == 0 {
		return nil, autocounter.ErrNoResults
	}

	var tables []autocounter.Table
	for _, item := range res.Result {
		if item.Database == nil {
			log.Printf("Table service: Available: Workspace %s: Unexpected item of non-database type", ws.ID)
			continue
		}

		if err := hasIDProperty(*item.Database, autocounter.DefaultTableParamName, notion.PropertyTypeNumber); err != nil {
			continue
		}

		t, err := autocounter.New(item.Database.ID, ws.ID)
		if err != nil {
			log.Printf("Table service: Available: Workspace %s: Couldn't compose a table: %s", ws.ID, err)
			continue
		}

		tables = append(tables, t)
	}

	return tables, nil
}

func hasIDProperty(t notion.Database, prop string, ptype notion.PropertyType) error {
	for n, p := range t.Properties {
		if n != prop {
			continue
		}

		if p.Type != ptype {
			continue
		}

		return nil
	}

	return autocounter.ErrIncompatibleTable
}

// ListAllActive tables for the provided Workspace ID.
func (t *Table) ListAllActive(ctx context.Context, workspaceID string) ([]autocounter.Table, error) {
	return t.s.ListAllActiveTables(ctx, workspaceID)
}

// Disable the Table from the Workspace by the table ID.
func (t *Table) Disable(ctx context.Context, wsID string, tableID string) error {
	return t.s.DisableTable(ctx, wsID, tableID)
}

// Fill the Table within provided Workspace with autoincrementing IDs.
func (t *Table) Fill(ctx context.Context, tableID string, ws autocounter.Workspace) error {
	if tableID == "" {
		return errors.New("table id is required")
	}

	table, err := t.FetchForWs(ctx, ws.ID, tableID)
	if err != nil {
		return fmt.Errorf("couldn't fetch table: %w", err)
	}

	notionCli, err := notion.NewFromWorkspace(ws)
	if err != nil {
		return fmt.Errorf("couldn't initialize notion api client: %s", err)
	}

	// fetch latest page number.
	res, err := notionCli.QueryDatabase(ctx, tableID, notion.DBQueryReq{
		Filter: &notion.DBFilter{
			Property: table.ParamName,
			Number: &notion.DBFilterNumber{
				IsNotEmpty: true,
			},
		},
		Sorts: []notion.DBSort{{
			Property:  table.ParamName,
			Direction: notion.DBSortDirectionDesc,
		}},
		PageSize: 1,
	})
	switch {
	case err == autocounter.ErrIncompatibleTable:
		return t.Disable(ctx, ws.ID, tableID)
	case err == autocounter.ErrTableNotFound:
		return t.Disable(ctx, ws.ID, tableID)
	case err != nil:
		return err
	}

	// default value of the counter is 0.
	var counter int64
	if len(res.Result) != 0 {
		p, ok := res.Result[0].Properties[table.ParamName]
		if !ok {
			return fmt.Errorf("%w: missing column: %s", autocounter.ErrInvalidTableParam, table.ParamName)
		}
		if p.Type != notion.PropertyTypeNumber {
			return fmt.Errorf("%w: wrong type of column %s: %s", autocounter.ErrInvalidTableParam, table.ParamName, p.Type)
		}
		if p.Number == nil {
			return errors.New("unexpected empty column value")
		}
		counter = int64(*p.Number)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// for loop until context is cancelled.
	var cursor string
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// fetch batch of the pages IDs with empty number ordered by asc created at.
		res, err := notionCli.QueryDatabase(ctx, tableID, notion.DBQueryReq{
			StartCursor: cursor,
			Filter: &notion.DBFilter{
				Property: table.ParamName,
				Number: &notion.DBFilterNumber{
					IsEmpty: true,
				},
			},
			Sorts: []notion.DBSort{{
				Timestamp: notion.DBSortTimestampCreated,
				Direction: notion.DBSortDirectionAsc,
			}},
		})
		if err != nil {
			return fmt.Errorf("couldn't fetch next batch of pages from db %s: %s", tableID, err)
		}

		// split into chunks
		chunks := chunkSlice(res.Result, defaultUpdateChunkSize)
		for _, ch := range chunks {
			wg := &sync.WaitGroup{}
			for _, p := range ch {
				wg.Add(1)

				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				counter++
				go func(num float64, pageID string, done func()) {
					defer done()

					_, err := notionCli.PatchPage(ctx, pageID, notion.PatchPageReq{
						Properties: map[string]notion.PageProperty{
							table.ParamName: {
								Type:   "number",
								Number: &num,
							},
						},
					})
					if err != nil {
						log.Printf("error: %s", err)
						cancel()
					}
				}(float64(counter), p.ID, wg.Done)
			}

			wg.Wait()
		}
		if !res.HasMore {
			return nil
		}

		cursor = *res.NextCursor
	}
}

func chunkSlice(slice []notion.Page, chunkSize int) [][]notion.Page {
	var chunks [][]notion.Page
	for {
		if len(slice) == 0 {
			break
		}

		// necessary check to avoid slicing beyond
		// slice capacity
		if len(slice) < chunkSize {
			chunkSize = len(slice)
		}

		chunks = append(chunks, slice[0:chunkSize])
		slice = slice[chunkSize:]
	}

	return chunks
}

// NonActiveDiff returns a list of tables that aren't registered or not active for the autofill.
// The new tables normally appear if customer decides to observe a new table,
// or at the first time the workspace is registered.
func (t *Table) NonActiveDiff(ctx context.Context, ws autocounter.Workspace) ([]autocounter.Table, error) {
	// fetch all the available tables from the tenant.
	tables, err := t.Available(ctx, ws)
	if err != nil {
		return nil, err
	}

	var tableIDs []string
	for _, t := range tables {
		tableIDs = append(tableIDs, t.ID)
	}

	// find the subset of active tables.
	reg, err := t.Active(ctx, ws.ID, tableIDs)
	if err != nil {
		return nil, err
	}

	if len(reg) == len(tableIDs) {
		return nil, err
	}

	notRegIDs := autocounter.Diff(tableIDs, reg)
	mapNotRegIDs := map[string]struct{}{}
	for _, id := range notRegIDs {
		mapNotRegIDs[id] = struct{}{}
	}

	var nonRegTables []autocounter.Table
	for _, t := range tables {
		if _, ok := mapNotRegIDs[t.ID]; !ok {
			continue
		}

		table := t

		nonRegTables = append(nonRegTables, table)
	}
	return nonRegTables, nil
}

// RegisterConc registers the provided table in a concurrent manner.
func (t *Table) RegisterConc(ctx context.Context, ts []autocounter.Table) {
	wg := &sync.WaitGroup{}
	for _, at := range ts {
		wg.Add(1)

		go func(at autocounter.Table) {
			defer wg.Done()
			if err := t.Register(ctx, at.WorkspaceID, at); err != nil {
				log.Printf("ScanWorker: Workspace %s: Couldn't register a table %s: %s", at.WorkspaceID, at.ID, err)
			}
		}(at)
	}
	wg.Wait()
}

// ProcWs in concurrent manner.
func (t *Table) ProcWs(ctx context.Context, ws autocounter.Workspace) (autocounter.Workspace, error) {
	ts, err := t.ListAllActive(ctx, ws.ID)
	switch {
	case err == autocounter.ErrNoResults:
	case err != nil:
		return autocounter.Workspace{}, fmt.Errorf("workspace %s: couldn't process tables: %w", ws.ID, err)
	}

	// parallelize the table fill.
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		ts, err := t.NonActiveDiff(ctx, ws)
		if err != nil {
			log.Printf("Table service: workspace %s: couldn't fetch unregistered tables: %s", ws.ID, err)
			return
		}
		t.RegisterConc(ctx, ts)
	}()

	for _, tt := range ts {
		wg.Add(1)
		go func(tID string) {
			defer wg.Done()
			if err := t.Fill(ctx, tID, ws); err != nil {
				log.Printf("RunWorker: workspace %s: couldn't fill the table %s: %s", ws.ID, tID, err)
			}
		}(tt.ID)
	}
	wg.Wait()

	return ws, nil
}
