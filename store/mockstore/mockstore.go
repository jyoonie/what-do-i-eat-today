package mockstore

import "wdiet/store"

var _ store.Store = (*Mockstore)(nil)

type Mockstore struct {
	PingOverride func() error
}

func (m *Mockstore) Ping() error {
	if m.PingOverride != nil {
		return m.PingOverride()
	}
	return nil
}
