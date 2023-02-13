package service

import (
	"time"

	"github.com/google/uuid"
)

//authentication handler에서 email이랑 password만 따로 받아서 json으로 주고받는 login 모델을 만들어야겠찌?
//밑에 User struct에서 굳이 password 필드를 user한테 받자고 남겨두는 것보단.. 왜냐면 response에서는 이 password 필드를 항상 empty로 둘거니께..

type User struct {
	UserUUID uuid.UUID `json:"user_uuid,omitempty"`
	//HashedPassword string    `json:"hashed_password,omitempty"`
	Active       bool   `json:"active,omitempty"`
	FirstName    string `json:"first_name,omitempty"`
	LastName     string `json:"last_name,omitempty"`
	EmailAddress string `json:"email_address,omitempty"`
	//CreatedAt    time.Time `json:"created_at,omitempty"`
	//UpdatedAt    time.Time `json:"updated_at,omitempty"`
}

type Login struct {
	EmailAddress string `json:"email_address,omitempty"`
	Password     string `json:"password,omitempty"`
}

type Token struct {
	Token string `json:"token,omitempty"`
}

type Ingredient struct {
	IngredientUUID uuid.UUID `json:"ingredient_uuid,omitempty"`
	IngredientName string    `json:"ingredient_name,omitempty"`
	Category       string    `json:"category,omitempty"`
	DaysUntilExp   int       `json:"days_until_exp,omitempty"`
	//created_at           time.Time
	//updated_at           time.Time
}

type SearchIngredient struct {
	IngredientName string `json:"ingredient_name,omitempty"`
	Category       string `json:"category,omitempty"`
}

type FridgeIngredient struct {
	UserUUID       uuid.UUID `json:"user_uuid,omitempty"`
	IngredientUUID uuid.UUID `json:"ingredient_uuid,omitempty"`
	Amount         int       `json:"amount,omitempty"`
	Unit           string    `json:"unit,omitempty"`
	PurchasedDate  time.Time `json:"purchased_date,omitempty"`
	ExpirationDate time.Time `json:"expiration_date,omitempty"`
	// created_at       time.Time
	// updated_at       time.Time
}

type DeleteFIngr struct {
	UserUUID       uuid.UUID `json:"user_uuid,omitempty"`
	IngredientUUID uuid.UUID `json:"ingredient_uuid,omitempty"`
}

type Recipe struct {
	RecipeUUID   uuid.UUID           `json:"recipe_uuid,omitempty"`
	UserUUID     uuid.UUID           `json:"user_uuid,omitempty"`
	RecipeName   string              `json:"recipe_name,omitempty"`
	Category     string              `json:"category,omitempty"`
	Ingredients  []RecipeIngredient  `json:"ingredients,omitempty"`  //여기에서 받은건 recipe ingredients 테이블에 저장된당
	Instructions []RecipeInstruction `json:"instructions,omitempty"` //여기에서 받은건 recipe instructions 테이블에 저장된당
	// CreatedAt  time.Time `json:"created_at,omitempty"`
	// UpdatedAt  time.Time `json:"updated_at,omitempty"`
}

type SearchRecipes struct {
	UserUUID   uuid.UUID `json:"user_uuid,omitempty"`
	RecipeName string    `json:"recipe_name,omitempty"`
	Category   string    `json:"category,omitempty"`
}

type RecipeIngredient struct {
	IngredientUUID uuid.UUID `json:"ingredient_uuid,omitempty"`
	Amount         int       `json:"amount,omitempty"`
	Unit           string    `json:"unit,omitempty"`
}

type RecipeInstruction struct {
	StepNum     int    `json:"step_num,omitempty"`
	Instruction string `json:"instruction,omitempty"`
}
