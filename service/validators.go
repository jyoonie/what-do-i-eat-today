package service

import (
	"github.com/google/uuid"
)

func isValidLoginRequest(l Login) bool {
	if l.EmailAddress == "" || l.Password == "" {
		return false
	}

	return true
}

func isValidCreateUserRequest(u User, pwd string) bool {
	switch {
	case u.UserUUID != uuid.Nil:
		return false
	case u.FirstName == "":
		return false
	case u.LastName == "":
		return false
	case u.EmailAddress == "":
		return false
	case pwd == "":
		return false
	}

	return true
}

func isValidUpdateUserRequest(u User, uidFromPath uuid.UUID) bool {
	switch {
	case uidFromPath != u.UserUUID:
		return false
	case u.UserUUID == uuid.Nil: //when you access a field on a pointer, go has to dereference the pointer. If you pass in a nil there, you'll panic, because you're dereferencing a nil pointer.(that's why you want to return *struct from a db method too.)
		return false
	case u.FirstName == "":
		return false
	case u.LastName == "":
		return false
	case u.EmailAddress == "":
		return false
	}

	return true
}

func isValidSearchIngrRequest(i SearchIngredient) bool {
	if i.IngredientName == "" && i.Category == "" {
		return false
	}

	return true
}

func isValidCreateIngrRequest(i Ingredient) bool {
	switch {
	case i.IngredientUUID != uuid.Nil:
		return false
	case i.IngredientName == "":
		return false
	case !isCategory(i.Category):
		return false
	case i.DaysUntilExp < 0:
		return false
	}

	return true
}

func isCategory(c string) bool {
	switch {
	case c == "vegetables":
		return true
	case c == "fruits":
		return true
	case c == "meat":
		return true
	case c == "fish":
		return true
	case c == "eggs":
		return true
	case c == "dairy":
		return true
	case c == "grains":
		return true
	case c == "water":
		return true
	case c == "etc":
		return true
	}

	return false
}

func isValidUpdateIngrRequest(i Ingredient, uidFromPath uuid.UUID) bool {
	switch {
	case uidFromPath != i.IngredientUUID:
		return false
	case i.IngredientUUID == uuid.Nil:
		return false
	case i.IngredientName == "":
		return false
	case !isCategory(i.Category):
		return false
	case i.DaysUntilExp < 0:
		return false
	}

	return true
}

func isValidCreateFIngrRequest(f FridgeIngredient) bool {
	switch {
	case f.UserUUID == uuid.Nil:
		return false
	case f.IngredientUUID == uuid.Nil:
		return false
	case f.Amount <= 0:
		return false
	case f.Unit == "":
		return false
	case f.PurchasedDate.IsZero():
		return false
	case !f.ExpirationDate.IsZero():
		return false
	}

	return true
}

func isValidUpdateFIngrRequest(f FridgeIngredient, uidFromPath uuid.UUID) bool {
	switch {
	case uidFromPath != f.IngredientUUID:
		return false
	case f.UserUUID == uuid.Nil:
		return false
	case f.IngredientUUID == uuid.Nil:
		return false
	case f.Amount <= 0:
		return false
	case f.Unit == "":
		return false
	case f.PurchasedDate.IsZero():
		return false
	case !f.ExpirationDate.IsZero(): //always validate the data like the front end is retarded
		return false
	}

	return true
}

// func isValidDeleteFIngrRequest(f DeleteFIngr, uidFromPath uuid.UUID) bool { //이제 user_uuid랑 ingredient_uuid 둘 다 파람으로 넘겨줘서 필요없어짐 ㅋ
// 	switch {
// 	case uidFromPath != f.UserUUID:
// 		return false
// 	case f.UserUUID == uuid.Nil:
// 		return false
// 	case f.IngredientUUID == uuid.Nil:
// 		return false
// 	}

// 	return true
// }

func isValidSearchRecipesRequest(r SearchRecipes) bool {
	if r.UserUUID == uuid.Nil && r.RecipeName == "" && r.Category == "" {
		return false
	}

	return true
}

func isValidCreateRecipeRequest(r Recipe) bool {
	switch {
	case r.RecipeUUID != uuid.Nil:
		return false
	case r.UserUUID == uuid.Nil:
		return false
	case r.RecipeName == "":
		return false
	case r.Category == "":
		return false
	case len(r.Ingredients) == 0:
		return false
	case len(r.Instructions) == 0:
		return false
	}

	return true
}

func isValidUpdateRecipeRequest(r Recipe, uidFromPath uuid.UUID) bool {
	switch {
	case uidFromPath != r.RecipeUUID:
		return false
	case r.RecipeUUID == uuid.Nil:
		return false
	case r.UserUUID == uuid.Nil:
		return false
	case r.RecipeName == "":
		return false
	case r.Category == "":
		return false
	case len(r.Ingredients) == 0:
		return false
	case len(r.Instructions) == 0:
		return false
	}

	return true
}
