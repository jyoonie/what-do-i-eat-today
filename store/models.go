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

type Login struct {
	EmailAddress string
	Password     string
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
