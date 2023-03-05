package inmemcache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	autocounter "github.com/notionplusid/core/app"
	m "github.com/notionplusid/core/app/internal/mock"
)

func TestClient(t *testing.T) {
	t.Run("returns error if no storage passed", func(t *testing.T) {
		_, err := New(nil)
		assert.Error(t, err)
	})

	t.Run("with storage provided", func(t *testing.T) {
		s := &m.Storage{}

		client, err := New(s)
		assert.NoError(t, err)

		t.Run("scenario 1", func(t *testing.T) {
			now := time.Now()
			ctx := context.TODO()

			mws := []autocounter.Workspace{
				{
					ID:          "1",
					Token:       "123",
					ProcessedAt: now,
					CreatedAt:   now,
					UpdatedAt:   now,
				}, {
					ID:          "2",
					Token:       "456",
					ProcessedAt: now,
					CreatedAt:   now,
					UpdatedAt:   now,
				},
			}

			s.On("Workspaces", ctx).Return(mws, nil)

			mt := []autocounter.Table{
				{
					ID:          "1",
					WorkspaceID: "1",
					Status:      autocounter.StatusActive,
					ParamName:   "PlusID",
					CreatedAt:   now,
					UpdatedAt:   now,
				},
				{
					ID:          "2",
					WorkspaceID: "2",
					Status:      autocounter.StatusActive,
					ParamName:   "PlusID",
					CreatedAt:   now,
					UpdatedAt:   now,
				},
			}

			s.On("Tables", ctx).Return(mt, nil)

			t.Log("starting with sync")
			err := client.Sync(ctx)
			assert.NoError(t, err)

			t.Log("checking if workspace is in")
			ws, err := client.Workspace(ctx, mws[0].ID)
			assert.NoError(t, err)

			assert.Equal(t, mws[0], ws)

			storedWs := autocounter.Workspace{
				ID:          "3",
				Token:       "123",
				ProcessedAt: now,
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			s.On("StoreWorkspace", ctx, storedWs).Return(storedWs, nil)
			_, err = client.StoreWorkspace(ctx, storedWs)
			assert.NoError(t, err)

			ws, err = client.Workspace(ctx, storedWs.ID)
			assert.NoError(t, err)

			assert.Equal(t, storedWs, ws)

			wss, err := client.Workspaces(ctx)
			assert.NoError(t, err)
			assert.ElementsMatch(t, wss, append(mws, storedWs))
		})
	})
}
