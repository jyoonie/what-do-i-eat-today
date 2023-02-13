package store

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UserUUID       uuid.UUID
	HashedPassword string
	Active         bool
	FirstName      string
	LastName       string
	EmailAddress   string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Ingredient struct {
	IngredientUUID uuid.UUID
	IngredientName string
	Category       string
	DaysUntilExp   int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// type Query struct {
// 	IngredientName *string //I have to check if they're null. If they're not null, I'm going to add each quer to base sql query
// 	Category       *string
// 	DaysUntilExp   *int
// }

type SearchIngredient struct {
	IngredientName *string
	Category       *string
}

type FridgeIngredient struct {
	UserUUID       uuid.UUID
	IngredientUUID uuid.UUID
	Amount         int
	Unit           string
	PurchasedDate  time.Time
	ExpirationDate time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type DeleteFIngr struct {
	UserUUID       uuid.UUID
	IngredientUUID uuid.UUID
}

type Recipe struct {
	RecipeUUID   uuid.UUID
	UserUUID     uuid.UUID
	RecipeName   string
	Category     string
	Ingredients  []RecipeIngredient
	Instructions []RecipeInstruction
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type SearchRecipes struct {
	UserUUID   *uuid.UUID
	RecipeName *string
	Category   *string
}

type RecipeIngredient struct { //your db model always matches your table. 그래서 여기에서는 init magrate up에 있는 모든 필드 다 있음.
	RecipeUUID     uuid.UUID
	IngredientUUID uuid.UUID
	Amount         int
	Unit           string
}

type RecipeInstruction struct {
	RecipeUUID  uuid.UUID
	StepNum     int
	Instruction string
}
