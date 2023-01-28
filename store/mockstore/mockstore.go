package mockstore

import (
	"context"
	"time"
	"wdiet/store"

	"github.com/google/uuid"
)

var _ store.Store = (*Mockstore)(nil)

type Mockstore struct {
	PingOverride       func() error
	GetUserOverride    func(ctx context.Context, id uuid.UUID) (*store.User, error)
	CreateUserOverride func(ctx context.Context, u store.User) (*store.User, error)
	UpdateUserOverride func(ctx context.Context, u store.User) (*store.User, error)
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
	return &store.User{
		UserUUID:       uuid.MustParse("080b5f09-527b-4581-bb56-19adbfe50ebf"),
		HashedPassword: "5994471abb01112afcc18159f6cc74b4f511b99806da59b3caf5a9c173cacfc5",
		Active:         true,
		FirstName:      "jy",
		LastName:       "woo",
		EmailAddress:   "jywoo92324@gmail.com"}, nil
}

func (m *Mockstore) CreateUser(ctx context.Context, u store.User) (*store.User, error) {
	if m.CreateUserOverride != nil {
		return m.CreateUserOverride(ctx, u)
	}

	return &store.User{
		UserUUID:       uuid.New(),
		HashedPassword: u.HashedPassword,
		Active:         u.Active,
		FirstName:      u.FirstName,
		LastName:       u.LastName,
		EmailAddress:   u.EmailAddress,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}, nil
}

func (m *Mockstore) UpdateUser(ctx context.Context, u store.User) (*store.User, error) {
	if m.UpdateUserOverride != nil {
		return m.UpdateUserOverride(ctx, u)
	}

	u.UpdatedAt = time.Now()

	return &u, nil //method.go의 UpdateUser를 따라 updated_at 필드를 now()로 바꿔줌. 이게 best practice. 근데 그 필드는 무시하고 request 그대로 반환해도 그게 그거다. mock store는 너무 빡빡하게 굴지말자 ^_^;
}
