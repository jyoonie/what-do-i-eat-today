package mockstore

import (
	"context"
	"time"
	"wdiet/store"

	"github.com/google/uuid"
)

var _ store.Store = (*Mockstore)(nil)

type Mockstore struct {
	PingOverride              func() error
	GetUserOverride           func(ctx context.Context, id uuid.UUID) (*store.User, error)
	CreateUserOverride        func(ctx context.Context, u store.User) (*store.User, error)
	UpdateUserOverride        func(ctx context.Context, u store.User) (*store.User, error)
	GetIngredientOverride     func(ctx context.Context, id uuid.UUID) (*store.Ingredient, error)
	CreateIngredientOverride  func(ctx context.Context, i store.Ingredient) (*store.Ingredient, error)
	UpdateIngredientOverride  func(ctx context.Context, i store.Ingredient) (*store.Ingredient, error)
	DeleteIngredientOverride  func(ctx context.Context, id uuid.UUID) error
	SearchIngredientsOverride func(ctx context.Context, i store.Ingredient) ([]store.Ingredient, error)
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

func (m *Mockstore) GetIngredient(ctx context.Context, id uuid.UUID) (*store.Ingredient, error) {
	if m.GetIngredientOverride != nil {
		return m.GetIngredientOverride(ctx, id)
	}
	return &store.Ingredient{
		IngredientUUID: uuid.MustParse("080b5f09-527b-4581-bb56-19adbfe50ebf"),
		IngredientName: "onion",
		Category:       "vegetables",
		DaysUntilExp:   7,
	}, nil
}

func (m *Mockstore) CreateIngredient(ctx context.Context, i store.Ingredient) (*store.Ingredient, error) {
	if m.CreateIngredientOverride != nil {
		return m.CreateIngredientOverride(ctx, i)
	}

	return &store.Ingredient{
		IngredientUUID: uuid.New(),
		IngredientName: i.IngredientName,
		Category:       i.Category,
		DaysUntilExp:   i.DaysUntilExp,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}, nil
}

func (m *Mockstore) UpdateIngredient(ctx context.Context, i store.Ingredient) (*store.Ingredient, error) {
	if m.UpdateIngredientOverride != nil {
		return m.UpdateIngredientOverride(ctx, i)
	}

	i.UpdatedAt = time.Now()

	return &i, nil
}

func (m *Mockstore) DeleteIngredient(ctx context.Context, id uuid.UUID) error {
	if m.DeleteIngredientOverride != nil {
		return m.DeleteIngredientOverride(ctx, id)
	}

	return nil
}

func (m *Mockstore) SearchIngredients(ctx context.Context, i store.Ingredient) ([]store.Ingredient, error) {
	if m.SearchIngredientsOverride != nil {
		return m.SearchIngredientsOverride(ctx, i)
	}

	return []store.Ingredient{
		{
			IngredientUUID: uuid.MustParse("080b5f09-527b-4581-bb56-19adbfe50ebf"),
			IngredientName: "tuna",
			Category:       "tuna kimbap",
			DaysUntilExp:   3,
		},
		{
			IngredientUUID: uuid.MustParse("080b5f09-527b-4581-bb56-19adbfe50ebf"),
			IngredientName: "tuna",
			Category:       "tuna sushi",
			DaysUntilExp:   3,
		},
	}, nil
}
