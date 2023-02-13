package service

import (
	"wdiet/store"

	"github.com/google/uuid"
)

func apiUser2DBUser(u User) store.User { //your user always submits a value, that's why you take the value. 이 모델은 핸들러에서 요청 받은거 db model로 변환할 때 쓴다.
	return store.User{
		UserUUID:     u.UserUUID, //for every user related request, this field is needed, but for createuser request, you don't need this field. So you can just leave it empty.
		Active:       u.Active,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		EmailAddress: u.EmailAddress,
	}
}

func dbUser2ApiUser(u *store.User) User { //your db always spits out the pointer, that's why you take the pointer.
	return User{
		UserUUID:     u.UserUUID,
		Active:       u.Active,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		EmailAddress: u.EmailAddress,
	}
}

func apiIngr2DBIngr(i Ingredient) store.Ingredient {
	return store.Ingredient{
		IngredientUUID: i.IngredientUUID,
		IngredientName: i.IngredientName,
		Category:       i.Category,
		DaysUntilExp:   i.DaysUntilExp,
	}
}

func dbIngr2ApiIngr(i *store.Ingredient) Ingredient {
	return Ingredient{
		IngredientUUID: i.IngredientUUID,
		IngredientName: i.IngredientName,
		Category:       i.Category,
		DaysUntilExp:   i.DaysUntilExp,
	}
}

func apiSearchIngr2DBSearchIngr(i SearchIngredient) store.SearchIngredient {
	var out store.SearchIngredient

	if i.IngredientName != "" {
		out.IngredientName = &i.IngredientName
	}
	if i.Category != "" {
		out.Category = &i.Category
	}
	return out
}

func apiFIngr2DBFIngr(f FridgeIngredient) store.FridgeIngredient { //api model에 있는 필드만 신경써.
	return store.FridgeIngredient{
		UserUUID:       f.UserUUID,
		IngredientUUID: f.IngredientUUID,
		Amount:         f.Amount,
		Unit:           f.Unit,
		PurchasedDate:  f.PurchasedDate,
		ExpirationDate: f.ExpirationDate,
	}
}

func dbFIngr2ApiFIngr(f *store.FridgeIngredient) FridgeIngredient {
	return FridgeIngredient{
		UserUUID:       f.UserUUID,
		IngredientUUID: f.IngredientUUID,
		Amount:         f.Amount,
		Unit:           f.Unit,
		PurchasedDate:  f.PurchasedDate,
		ExpirationDate: f.ExpirationDate,
	}
}

// func apiDeleteFI2DBDeleteFI(f DeleteFIngr) store.DeleteFIngr { //이제 user_uuid랑 ingredient_uuid 둘 다 파람으로 넘겨줘서 필요없어짐 ㅋ
// 	return store.DeleteFIngr{
// 		UserUUID:       f.UserUUID,
// 		IngredientUUID: f.IngredientUUID,
// 	}
// }

func apiRIngr2DBRIngr(r []RecipeIngredient) []store.RecipeIngredient {
	var ingredients []store.RecipeIngredient

	for _, ingr := range r {
		var s store.RecipeIngredient

		s.IngredientUUID = ingr.IngredientUUID
		s.Amount = ingr.Amount
		s.Unit = ingr.Unit
		ingredients = append(ingredients, s)
	}

	return ingredients
}

func apiRInst2DBRInst(r []RecipeInstruction) []store.RecipeInstruction {
	var instructions []store.RecipeInstruction

	for _, inst := range r {
		var s store.RecipeInstruction

		s.StepNum = inst.StepNum
		s.Instruction = inst.Instruction
		instructions = append(instructions, s)
	}

	return instructions
}

func apiRecipe2DBRecipe(r Recipe) store.Recipe {
	return store.Recipe{
		RecipeUUID:   r.RecipeUUID,
		UserUUID:     r.UserUUID,
		RecipeName:   r.RecipeName,
		Category:     r.Category,
		Ingredients:  apiRIngr2DBRIngr(r.Ingredients),
		Instructions: apiRInst2DBRInst(r.Instructions),
	}
}

func DBRIngr2apiRIngr(r []store.RecipeIngredient) []RecipeIngredient {
	var ingredients []RecipeIngredient

	for _, ingr := range r {
		var s RecipeIngredient

		s.IngredientUUID = ingr.IngredientUUID
		s.Amount = ingr.Amount
		s.Unit = ingr.Unit
		ingredients = append(ingredients, s)
	}

	return ingredients
}

func DBRInst2apiRInst(r []store.RecipeInstruction) []RecipeInstruction {
	var instructions []RecipeInstruction

	for _, inst := range r {
		var s RecipeInstruction

		s.StepNum = inst.StepNum
		s.Instruction = inst.Instruction
		instructions = append(instructions, s)
	}

	return instructions
}

func dbRecipe2ApiRecipe(r *store.Recipe) Recipe {
	return Recipe{
		RecipeUUID:   r.RecipeUUID,
		UserUUID:     r.UserUUID,
		RecipeName:   r.RecipeName,
		Category:     r.Category,
		Ingredients:  DBRIngr2apiRIngr(r.Ingredients),
		Instructions: DBRInst2apiRInst(r.Instructions),
	}
}

func apiSearchR2DBSearchR(r SearchRecipes) store.SearchRecipes {
	var out store.SearchRecipes

	if r.UserUUID != uuid.Nil {
		out.UserUUID = &r.UserUUID
	}
	if r.RecipeName != "" {
		out.RecipeName = &r.RecipeName
	}
	if r.Category != "" {
		out.Category = &r.Category
	}

	return out
}
