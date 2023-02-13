package mockstore

import (
	"context"
	"fmt"
	"time"
	"wdiet/store"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var _ store.Store = (*Mockstore)(nil)

type Mockstore struct {
	PingOverride func() error

	GetUserOverride        func(ctx context.Context, id uuid.UUID) (*store.User, error)
	GetUserByEmailOverride func(ctx context.Context, email string) (*store.User, error)
	CreateUserOverride     func(ctx context.Context, u store.User) (*store.User, error)
	UpdateUserOverride     func(ctx context.Context, u store.User) (*store.User, error)

	GetIngredientOverride     func(ctx context.Context, id uuid.UUID) (*store.Ingredient, error)
	SearchIngredientsOverride func(ctx context.Context, i store.SearchIngredient) ([]store.Ingredient, error)
	CreateIngredientOverride  func(ctx context.Context, i store.Ingredient) (*store.Ingredient, error)
	UpdateIngredientOverride  func(ctx context.Context, i store.Ingredient) (*store.Ingredient, error)
	DeleteIngredientOverride  func(ctx context.Context, id uuid.UUID) error

	ListFridgeIngredientsOverride  func(ctx context.Context, id uuid.UUID) ([]store.FridgeIngredient, error)
	CreateFridgeIngredientOverride func(ctx context.Context, f store.FridgeIngredient) (*store.FridgeIngredient, error)
	UpdateFridgeIngredientOverride func(ctx context.Context, f store.FridgeIngredient) (*store.FridgeIngredient, error)
	DeleteFridgeIngredientOverride func(ctx context.Context, uid uuid.UUID, fid uuid.UUID) error

	GetRecipeOverride     func(ctx context.Context, id uuid.UUID) (*store.Recipe, error)
	ListRecipesOverride   func(ctx context.Context, id uuid.UUID) ([]store.Recipe, error)
	SearchRecipesOverride func(ctx context.Context, r store.SearchRecipes) ([]store.Recipe, error)
	CreateRecipeOverride  func(ctx context.Context, r store.Recipe) (*store.Recipe, error)
	UpdateRecipeOverride  func(ctx context.Context, r store.Recipe) (*store.Recipe, error)
	DeleteRecipeOverride  func(ctx context.Context, id uuid.UUID) error
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
		UserUUID:       id,
		HashedPassword: "5994471abb01112afcc18159f6cc74b4f511b99806da59b3caf5a9c173cacfc5",
		Active:         true,
		FirstName:      "jy",
		LastName:       "woo",
		EmailAddress:   "jywoo92324@gmail.com",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}, nil
}

func (m *Mockstore) GetUserByEmail(ctx context.Context, email string) (*store.User, error) {
	if m.GetUserByEmailOverride != nil {
		return m.GetUserByEmailOverride(ctx, email)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("hello"), bcrypt.MinCost)
	if err != nil {
		return nil, fmt.Errorf("error hashing mockstore default password: %w", err)
	}

	return &store.User{
		UserUUID:       uuid.MustParse("080b5f09-527b-4581-bb56-19adbfe50ebf"),
		HashedPassword: string(hashedPassword),
		Active:         false,
		FirstName:      "jy",
		LastName:       "woo",
		EmailAddress:   email, //"jywoo92324@gmail.com"
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}, nil
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
		IngredientUUID: id,
		IngredientName: "onion",
		Category:       "vegetables",
		DaysUntilExp:   7,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}, nil
}

func (m *Mockstore) SearchIngredients(ctx context.Context, i store.SearchIngredient) ([]store.Ingredient, error) {
	if m.SearchIngredientsOverride != nil {
		return m.SearchIngredientsOverride(ctx, i)
	}

	return []store.Ingredient{
		{
			IngredientUUID: uuid.MustParse("080b5f09-527b-4581-bb56-19adbfe50ebf"),
			IngredientName: "tuna",
			Category:       "tuna kimbap",
			DaysUntilExp:   3,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			IngredientUUID: uuid.MustParse("080b5f09-527b-4581-bb56-19adbfe50ebf"),
			IngredientName: "tuna",
			Category:       "tuna sushi",
			DaysUntilExp:   3,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
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

func (m *Mockstore) ListFridgeIngredients(ctx context.Context, id uuid.UUID) ([]store.FridgeIngredient, error) {
	if m.ListFridgeIngredientsOverride != nil {
		return m.ListFridgeIngredientsOverride(ctx, id)
	}

	return []store.FridgeIngredient{
		{
			UserUUID:       id,
			IngredientUUID: uuid.MustParse("ffff7c73-52b0-4e3d-bf3f-0c26785ef972"),
			Amount:         3,
			Unit:           "kg",
			PurchasedDate:  time.Date(2023, time.March, 24, 15, 0, 0, 0, time.Now().Location()),
			ExpirationDate: time.Date(2023, time.March, 24, 15, 0, 0, 0, time.Now().Location()),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			UserUUID:       id,
			IngredientUUID: uuid.MustParse("2c98fff4-7ccc-4536-8259-67a88380e99c"),
			Amount:         2,
			Unit:           "L",
			PurchasedDate:  time.Date(2023, time.March, 24, 15, 0, 0, 0, time.Now().Location()),
			ExpirationDate: time.Date(2023, time.March, 31, 15, 0, 0, 0, time.Now().Location()),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
	}, nil
}

func (m *Mockstore) CreateFridgeIngredient(ctx context.Context, f store.FridgeIngredient) (*store.FridgeIngredient, error) {
	if m.CreateFridgeIngredientOverride != nil {
		return m.CreateFridgeIngredientOverride(ctx, f)
	}

	return &store.FridgeIngredient{
		UserUUID:       f.UserUUID,
		IngredientUUID: f.IngredientUUID,
		Amount:         f.Amount,
		Unit:           f.Unit,
		PurchasedDate:  f.PurchasedDate,
		ExpirationDate: f.ExpirationDate,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}, nil
}

func (m *Mockstore) UpdateFridgeIngredient(ctx context.Context, f store.FridgeIngredient) (*store.FridgeIngredient, error) {
	if m.UpdateFridgeIngredientOverride != nil {
		return m.UpdateFridgeIngredientOverride(ctx, f)
	}

	f.UpdatedAt = time.Now()

	return &f, nil
}

func (m *Mockstore) DeleteFridgeIngredient(ctx context.Context, uid uuid.UUID, fid uuid.UUID) error {
	if m.DeleteFridgeIngredientOverride != nil {
		return m.DeleteFridgeIngredientOverride(ctx, uid, fid)
	}

	return nil
}

func (m *Mockstore) GetRecipe(ctx context.Context, id uuid.UUID) (*store.Recipe, error) {
	if m.GetRecipeOverride != nil {
		return m.GetRecipeOverride(ctx, id)
	}

	return &store.Recipe{
		RecipeUUID: id,
		UserUUID:   uuid.MustParse("2c98fff4-7ccc-4536-8259-67a88380e99c"),
		RecipeName: "kimchi fried rice",
		Category:   "Korean",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}, nil
}

func (m *Mockstore) ListRecipes(ctx context.Context, id uuid.UUID) ([]store.Recipe, error) {
	if m.ListRecipesOverride != nil {
		return m.ListRecipesOverride(ctx, id)
	}

	return []store.Recipe{
		{
			RecipeUUID: uuid.MustParse("ffff7c73-52b0-4e3d-bf3f-0c26785ef972"),
			UserUUID:   id,
			RecipeName: "kimchi jeon",
			Category:   "Korean",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			RecipeUUID: uuid.MustParse("2c98fff4-7ccc-4536-8259-67a88380e99c"),
			UserUUID:   id,
			RecipeName: "kimchi mandu",
			Category:   "Korean",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}, nil
}

func (m *Mockstore) SearchRecipes(ctx context.Context, r store.SearchRecipes) ([]store.Recipe, error) {
	if m.SearchRecipesOverride != nil {
		return m.SearchRecipesOverride(ctx, r)
	}

	return []store.Recipe{
		{
			RecipeUUID: uuid.MustParse("ffff7c73-52b0-4e3d-bf3f-0c26785ef972"),
			UserUUID:   uuid.MustParse("080b5f09-527b-4581-bb56-19adbfe50ebf"),
			RecipeName: "salmon nigiri",
			Category:   "Japanese",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			RecipeUUID: uuid.MustParse("2c98fff4-7ccc-4536-8259-67a88380e99c"),
			UserUUID:   uuid.MustParse("ebe96725-44ef-47ee-979f-8baf823d7283"),
			RecipeName: "salmon nigiri",
			Category:   "Japanese",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}, nil
}

func (m *Mockstore) CreateRecipe(ctx context.Context, r store.Recipe) (*store.Recipe, error) {
	if m.CreateRecipeOverride != nil {
		return m.CreateRecipeOverride(ctx, r)
	}

	recipeID := uuid.New()

	return &store.Recipe{
		RecipeUUID: recipeID,
		UserUUID:   uuid.MustParse("2c98fff4-7ccc-4536-8259-67a88380e99c"),
		RecipeName: "kimchi fried rice",
		Category:   "Korean",
		Ingredients: []store.RecipeIngredient{
			{
				RecipeUUID:     recipeID,                                               //이바부야.. methods.go에서 이미 recipe.RecipeUUID로 채워놨자나..
				IngredientUUID: uuid.MustParse("2c98fff4-7ccc-4536-8259-67a88380e99b"), //니 맘대로 uuid 바꿔도 상관없어~ 숫자, 소문자, 대문자 다 상관없어, as long as they're 0~9,a~f
				Amount:         1,
				Unit:           "kg",
			},
			{
				RecipeUUID:     recipeID,
				IngredientUUID: uuid.MustParse("2c98fff4-7ccc-4536-8259-67a88380e99b"),
				Amount:         1,
				Unit:           "unit",
			},
		},
		Instructions: []store.RecipeInstruction{
			{
				RecipeUUID:  recipeID,
				StepNum:     1,
				Instruction: "Chop kimchi, onion and pork belly",
			},
			{
				RecipeUUID:  recipeID,
				StepNum:     2,
				Instruction: "grill the pan and put some oil on it",
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (m *Mockstore) UpdateRecipe(ctx context.Context, r store.Recipe) (*store.Recipe, error) {
	if m.UpdateRecipeOverride != nil {
		return m.UpdateRecipeOverride(ctx, r)
	}

	r.UpdatedAt = time.Now()

	return &r, nil
}

func (m *Mockstore) DeleteRecipe(ctx context.Context, id uuid.UUID) error {
	if m.DeleteRecipeOverride != nil {
		return m.DeleteRecipeOverride(ctx, id)
	}

	return nil
}
