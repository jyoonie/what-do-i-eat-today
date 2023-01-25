package mockstore

import (
	"context"
	"wdiet/store"

	"github.com/google/uuid"
)

var _ store.Store = (*Mockstore)(nil)

type Mockstore struct {
	PingOverride    func() error
	GetUserOverride func(ctx context.Context, id uuid.UUID) (*store.User, error)
}

func (m *Mockstore) Ping() error {
	if m.PingOverride != nil {
		return m.PingOverride()
	}
	return nil
}

func (m *Mockstore) GetUser(ctx context.Context, id uuid.UUID) (*store.User, error) {
	if m.GetUserOverride != nil {
		return m.GetUserOverride(ctx, id)
	}
	return nil, nil
}
