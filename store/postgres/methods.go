package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"wdiet/store"

	"github.com/google/uuid"
)

func (pg *PG) Ping() error { //implementing the Store interface, nice to separate the postgres definition with the methods that fulfill the interface.
	return pg.db.Ping() //PG struct has db field in it, now this is reaching it. Every database has a ping function(just like queryrowcontext), just to make sure that the connection is up and working.
}

const defaultTimeout = 5 * time.Second

func (pg *PG) GetUser(ctx context.Context, id uuid.UUID) (*store.User, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var user store.User

	row := pg.db.QueryRowContext(ctx, sqlGetUser, id)
	if err := row.Scan(
		&user.UserUUID,
		&user.HashedPassword,
		&user.Active,
		&user.FirstName,
		&user.LastName,
		&user.EmailAddress,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	return &user, nil
}

func (pg *PG) GetUserByEmail(ctx context.Context, email string) (*store.User, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var user store.User

	row := pg.db.QueryRowContext(ctx, sqlGetUserByEmail, email)
	if err := row.Scan(
		&user.UserUUID,       //don't half fill a struct, if you're returning a *store.User, return every field. Make every field valid, instead of just filling two fields of it.
		&user.HashedPassword, //don't mess up the order on a Scan, cause it's gonna follow the order from sql.go
		&user.Active,
		&user.FirstName,
		&user.LastName,
		&user.EmailAddress,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	return &user, nil
}

func (pg *PG) CreateUser(ctx context.Context, u store.User) (*store.User, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	tx, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	var user store.User

	row := tx.QueryRowContext(ctx, sqlCreateUser, //여기서 sqlCreateUser를 부를 때, initial_migrations의 user_uuid uuid not null default gen_random_uuid()가 실행됨. 그래서 자동으로 user_uuid 필드에 채워짐. 참고로, 여기서는 row를 create하고 그 필드들을 arg로 받은 값들로 채우고, 나머지 필드들은 default로 채움.
		u.HashedPassword,
		u.Active,
		u.FirstName,
		u.LastName,
		u.EmailAddress,
	)

	if err = row.Scan( //여기선 위에서 생성된 row를 scan해서 user에 복붙?한다음 그 user의 주소를 return하는고지..
		&user.UserUUID,
		&user.HashedPassword,
		&user.Active,
		&user.FirstName,
		&user.LastName,
		&user.EmailAddress,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	return &user, nil
}

func (pg *PG) UpdateUser(ctx context.Context, u store.User) (*store.User, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	tx, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	var user store.User

	row := tx.QueryRowContext(ctx, sqlUpdateUser,
		u.Active,
		u.FirstName,
		u.LastName,
		u.EmailAddress,
		u.UserUUID,
	)

	if err = row.Scan( //여기선 위에서 생성된 row를 scan해서 user에 복붙?한다음 그 user의 주소를 return하는고지..
		&user.UserUUID,
		&user.HashedPassword,
		&user.Active,
		&user.FirstName,
		&user.LastName,
		&user.EmailAddress,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			tx.Rollback()
			return nil, store.ErrNotFound
		}
		tx.Rollback()
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	return &user, nil
}

func (pg *PG) GetIngredient(ctx context.Context, id uuid.UUID) (*store.Ingredient, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var ingredient store.Ingredient

	row := pg.db.QueryRowContext(ctx, sqlGetIngredient, id)
	if err := row.Scan(
		&ingredient.IngredientUUID,
		&ingredient.IngredientName,
		&ingredient.Category,
		&ingredient.DaysUntilExp,
		&ingredient.CreatedAt,
		&ingredient.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("error getting ingredient: %w", err)
	}

	return &ingredient, nil
}

func (pg *PG) SearchIngredients(ctx context.Context, i store.SearchIngredient) ([]store.Ingredient, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var wheres []string
	var vars []interface{}
	var count int

	if i.IngredientName != nil {
		count++
		wheres = append(wheres, fmt.Sprintf(" ingredient_name = $%d", count))
		vars = append(vars, i.IngredientName)
	}
	if i.Category != nil {
		count++
		wheres = append(wheres, fmt.Sprintf(" category = $%d", count))
		vars = append(vars, i.Category)
	}

	whereClause := strings.Join(wheres, " AND ") //만약 하나가 안온다 그럼 join 자체가 안되기 때문에 AND로 묶이지도 않나보네..?

	var ingredients []store.Ingredient

	fmt.Println("query is ", sqlsearchIngredients+" WHERE "+whereClause, "vars are ", vars)
	rows, err := pg.db.QueryContext(ctx, sqlsearchIngredients+" WHERE "+whereClause, vars...)
	if err != nil {
		return nil, fmt.Errorf("error searching ingredients: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ingredient store.Ingredient
		if err := rows.Scan(
			&ingredient.IngredientUUID,
			&ingredient.IngredientName,
			&ingredient.Category,
			&ingredient.DaysUntilExp,
			&ingredient.CreatedAt,
			&ingredient.UpdatedAt,
		); err != nil {
			// if errors.Is(err, sql.ErrNoRows) {
			// 	return nil, store.ErrNotFound
			// }
			return nil, fmt.Errorf("error searching ingredients: %w", err)
		}
		ingredients = append(ingredients, ingredient)
	}
	// if err := rows.Close(); err != nil {
	// 	return nil, err
	// }
	// if err := rows.Err(); err != nil {
	// 	return nil, err
	// }
	return ingredients, nil
}

func (pg *PG) CreateIngredient(ctx context.Context, i store.Ingredient) (*store.Ingredient, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	tx, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating ingredient: %w", err)
	}

	var ingredient store.Ingredient

	row := tx.QueryRowContext(ctx, sqlCreateIngredient,
		&i.IngredientName,
		&i.Category,
		&i.DaysUntilExp,
	)

	if err = row.Scan(
		&ingredient.IngredientUUID,
		&ingredient.IngredientName,
		&ingredient.Category,
		&ingredient.DaysUntilExp,
		&ingredient.CreatedAt,
		&ingredient.UpdatedAt,
	); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error creating ingredient: %w", err)
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error creating ingredient: %w", err)
	}

	return &ingredient, nil
}

func (pg *PG) UpdateIngredient(ctx context.Context, i store.Ingredient) (*store.Ingredient, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	tx, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating ingredient: %w", err)
	}

	var ingredient store.Ingredient

	row := tx.QueryRowContext(ctx, sqlUpdateIngredient,
		i.IngredientName,
		i.Category,
		i.DaysUntilExp,
		i.IngredientUUID,
	)

	if err = row.Scan(
		&ingredient.IngredientUUID,
		&ingredient.IngredientName,
		&ingredient.Category,
		&ingredient.DaysUntilExp,
		&ingredient.CreatedAt,
		&ingredient.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			tx.Rollback()
			return nil, store.ErrNotFound
		}
		tx.Rollback()
		return nil, fmt.Errorf("error updating ingredient: %w", err)
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error updating ingredient: %w", err)
	}

	return &ingredient, nil
}

func (pg *PG) DeleteIngredient(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel() //it defers cancel until it returns from an error(actually, until it returns from current function, even john doesn't know what will happen after you return without an error. Does cancel() execute?) So what does it do? It cancels the context everywhere, so you don't do extra work when you know it's gonna fail. Context is per request.

	//writing to or deleting from db, you should be in a transaction
	tx, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error deleting ingredient: %w", err)
	}

	res, err := tx.ExecContext(ctx, sqlDeleteIngredient, id)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error deleting ingredient: %w", err)
	}

	if affected, _ := res.RowsAffected(); affected != 1 { //you do this to prevent update when without WHERE clause, or if there's two rows that have the same WHERE condition I don't need to do it when it's queryRowContext. Only works with multiple rows.
		tx.Rollback()
		return fmt.Errorf("error deleting ingredient, rows affected is %d instead of 1", affected)
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return fmt.Errorf("error deleting ingredient: %w", err)
	}

	return nil
}

func (pg *PG) ListFridgeIngredients(ctx context.Context, id uuid.UUID) ([]store.FridgeIngredient, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var fridgeIngredients []store.FridgeIngredient

	rows, err := pg.db.QueryContext(ctx, sqlListFridgeIngredients, id)
	if err != nil {
		return nil, fmt.Errorf("error listing fridge ingredients: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var fridgeIngredient store.FridgeIngredient
		if err := rows.Scan(
			&fridgeIngredient.UserUUID,
			&fridgeIngredient.IngredientUUID,
			&fridgeIngredient.Amount,
			&fridgeIngredient.Unit,
			&fridgeIngredient.PurchasedDate,
			&fridgeIngredient.ExpirationDate,
			&fridgeIngredient.CreatedAt,
			&fridgeIngredient.UpdatedAt,
		); err != nil {
			// if errors.Is(err, sql.ErrNoRows) {
			// 	return nil, store.ErrNotFound
			// }
			return nil, fmt.Errorf("error listing fridge ingredients: %w", err)
		}
		fridgeIngredients = append(fridgeIngredients, fridgeIngredient)
	}

	return fridgeIngredients, nil
}

func (pg *PG) CreateFridgeIngredient(ctx context.Context, f store.FridgeIngredient) (*store.FridgeIngredient, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	tx, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating fridge ingredient: %w", err)
	}

	var fridgeIngredient store.FridgeIngredient

	row := tx.QueryRowContext(ctx, sqlCreateFridgeIngredient,
		&f.UserUUID,
		&f.IngredientUUID,
		&f.Amount,
		&f.Unit,
		&f.PurchasedDate,
		&f.ExpirationDate,
	)

	err = row.Scan(
		&fridgeIngredient.UserUUID,
		&fridgeIngredient.IngredientUUID,
		&fridgeIngredient.Amount,
		&fridgeIngredient.Unit,
		&fridgeIngredient.PurchasedDate,
		&fridgeIngredient.ExpirationDate,
		&fridgeIngredient.CreatedAt,
		&fridgeIngredient.UpdatedAt,
	)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error creating fridge ingredient: %w", err)
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error creating fridge ingredient: %w", err)
	}

	return &fridgeIngredient, nil
}

func (pg *PG) UpdateFridgeIngredient(ctx context.Context, f store.FridgeIngredient) (*store.FridgeIngredient, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	tx, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating fridge ingredient: %w", err)
	}

	var fridgeIngredient store.FridgeIngredient

	row := tx.QueryRowContext(ctx, sqlUpdateFridgeIngredient, //여기에 나온 순서는 sql.go에서 $로 순서매겨놓은 순서.
		&f.Amount,
		&f.Unit,
		&f.PurchasedDate,
		&f.UserUUID,
		&f.IngredientUUID,
	)

	if err = row.Scan( //여기에 나온 순서는 sql.go에서 RETURNING 뒤의 필드 순서.
		&fridgeIngredient.UserUUID,
		&fridgeIngredient.IngredientUUID,
		&fridgeIngredient.Amount,
		&fridgeIngredient.Unit,
		&fridgeIngredient.PurchasedDate,
		&f.ExpirationDate, //update fridge ingredient에서 f.purchased date로 산출된 expiration date를 여기에 넣어줘야함..
		&fridgeIngredient.CreatedAt,
		&fridgeIngredient.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			tx.Rollback()
			return nil, store.ErrNotFound
		}
		tx.Rollback()
		return nil, fmt.Errorf("error updating fridge ingredient: %w", err)
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error updating fridge ingredient: %w", err)
	}

	return &fridgeIngredient, nil
}

func (pg *PG) DeleteFridgeIngredient(ctx context.Context, uid uuid.UUID, fid uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel() //to make sure that the cancel function runs, otherwise I have to say it before every return cause by error

	tx, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error deleting fridge ingredient: %w", err)
	}

	res, err := tx.ExecContext(ctx, sqlDeleteFridgeIngredient, uid, fid)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error deleting fridge ingredient: %w", err)
	}

	if affected, _ := res.RowsAffected(); affected != 1 {
		tx.Rollback()
		return fmt.Errorf("error deleting fridge ingredient, rows affected %d instead of 1", affected)
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return fmt.Errorf("error deleting fridge ingredient: %w", err)
	}

	return nil
}

func (pg *PG) GetRecipe(ctx context.Context, id uuid.UUID) (*store.Recipe, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var recipe store.Recipe

	row := pg.db.QueryRowContext(ctx, sqlGetRecipe, id)
	if err := row.Scan(
		&recipe.RecipeUUID,
		&recipe.UserUUID,
		&recipe.RecipeName,
		&recipe.Category,
		&recipe.CreatedAt,
		&recipe.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("error getting recipe: %w", err)
	}

	ingredients, err := pg.db.QueryContext(ctx, sqlListRecipeIngr, recipe.RecipeUUID)
	if err != nil {
		return nil, fmt.Errorf("error listing recipe ingredients: %w", err)
	}
	defer ingredients.Close()

	for ingredients.Next() {
		var recipeIngr store.RecipeIngredient
		if err := ingredients.Scan(
			&recipeIngr.RecipeUUID,
			&recipeIngr.IngredientUUID,
			&recipeIngr.Amount,
			&recipeIngr.Unit,
		); err != nil {
			return nil, fmt.Errorf("error listing recipe ingredients: %w", err)
		}
		recipe.Ingredients = append(recipe.Ingredients, recipeIngr)
	}

	instructions, err := pg.db.QueryContext(ctx, sqlListRecipeInst, recipe.RecipeUUID)
	if err != nil {
		return nil, fmt.Errorf("error listing recipe instructions: %w", err)
	}
	defer instructions.Close()

	for instructions.Next() {
		var recipeInst store.RecipeInstruction
		if err := instructions.Scan(
			&recipeInst.RecipeUUID,
			&recipeInst.StepNum,
			&recipeInst.Instruction,
		); err != nil {
			return nil, fmt.Errorf("error listing recipe instructions: %w", err)
		}
		recipe.Instructions = append(recipe.Instructions, recipeInst)
	}

	return &recipe, nil
}

func (pg *PG) ListRecipes(ctx context.Context, id uuid.UUID) ([]store.Recipe, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var recipes []store.Recipe

	rows, err := pg.db.QueryContext(ctx, sqlListRecipes, id)
	if err != nil {
		return nil, fmt.Errorf("error listing recipes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var recipe store.Recipe
		if err := rows.Scan(
			&recipe.RecipeUUID,
			&recipe.UserUUID,
			&recipe.RecipeName,
			&recipe.Category,
			&recipe.CreatedAt,
			&recipe.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("error listing recipes: %w", err)
		}

		ingredients, err := pg.db.QueryContext(ctx, sqlListRecipeIngr, recipe.RecipeUUID)
		if err != nil {
			return nil, fmt.Errorf("error listing recipe ingredients: %w", err)
		}
		defer ingredients.Close()

		for ingredients.Next() {
			var recipeIngr store.RecipeIngredient
			if err := ingredients.Scan(
				&recipeIngr.RecipeUUID,
				&recipeIngr.IngredientUUID,
				&recipeIngr.Amount,
				&recipeIngr.Unit,
			); err != nil {
				return nil, fmt.Errorf("error listing recipe ingredients: %w", err)
			}
			recipe.Ingredients = append(recipe.Ingredients, recipeIngr)
		}

		instructions, err := pg.db.QueryContext(ctx, sqlListRecipeInst, recipe.RecipeUUID)
		if err != nil {
			return nil, fmt.Errorf("error listing recipe instructions: %w", err)
		}
		defer instructions.Close()

		for instructions.Next() {
			var recipeInst store.RecipeInstruction
			if err := instructions.Scan(
				&recipeInst.RecipeUUID,
				&recipeInst.StepNum,
				&recipeInst.Instruction,
			); err != nil {
				return nil, fmt.Errorf("error listing recipe instructions: %w", err)
			}
			recipe.Instructions = append(recipe.Instructions, recipeInst)
		}

		recipes = append(recipes, recipe)
	}

	return recipes, nil
}

func (pg *PG) SearchRecipes(ctx context.Context, r store.SearchRecipes) ([]store.Recipe, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var wheres []string
	var vars []interface{}
	var count int

	if r.UserUUID != nil {
		count++
		wheres = append(wheres, fmt.Sprintf(" user_uuid = $%d", count))
		vars = append(vars, r.UserUUID)
	}

	if r.RecipeName != nil {
		count++
		wheres = append(wheres, fmt.Sprintf(" recipe_name = $%d", count))
		vars = append(vars, r.RecipeName)
	}

	if r.Category != nil {
		count++
		wheres = append(wheres, fmt.Sprintf(" category = $%d", count))
		vars = append(vars, r.Category)
	}

	whereClause := strings.Join(wheres, " AND ")

	var recipes []store.Recipe

	fmt.Println("sql statement is")
	rows, err := pg.db.QueryContext(ctx, sqlsearchRecipes+" WHERE "+whereClause, vars...)
	if err != nil {
		return nil, fmt.Errorf("error searching recipes: %w", err)
	}

	for rows.Next() {
		var recipe store.Recipe
		if err = rows.Scan( //여기서 하는 짓은, recipe table에서 찾은 row를 각각 recipe struct의 필드에 맞게 할당해주는 거임.
			&recipe.RecipeUUID,
			&recipe.UserUUID,
			&recipe.RecipeName,
			&recipe.Category,
			&recipe.CreatedAt,
			&recipe.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("error searching recipes: %w", err)
		}

		ingredients, err := pg.db.QueryContext(ctx, sqlListRecipeIngr, recipe.RecipeUUID)
		if err != nil {
			return nil, fmt.Errorf("error listing recipe ingredients: %w", err)
		}
		defer ingredients.Close()

		for ingredients.Next() {
			var recipeIngr store.RecipeIngredient
			if err := ingredients.Scan(
				&recipeIngr.RecipeUUID,
				&recipeIngr.IngredientUUID,
				&recipeIngr.Amount,
				&recipeIngr.Unit,
			); err != nil {
				return nil, fmt.Errorf("error listing recipe ingredients: %w", err)
			}
			recipe.Ingredients = append(recipe.Ingredients, recipeIngr) //여기서 하는 짓은, recipe_ingredients table에서 찾은 row를 각각 recipe struct의 필드에 맞게 할당해주는 거임. & 써야하나..
		}

		instructions, err := pg.db.QueryContext(ctx, sqlListRecipeInst, recipe.RecipeUUID)
		if err != nil {
			return nil, fmt.Errorf("error listing recipe instructions: %w", err)
		}
		defer instructions.Close()

		for instructions.Next() {
			var recipeInst store.RecipeInstruction
			if err := instructions.Scan(
				&recipeInst.RecipeUUID,
				&recipeInst.StepNum,
				&recipeInst.Instruction,
			); err != nil {
				return nil, fmt.Errorf("error listing recipe instructions: %w", err)
			}
			recipe.Instructions = append(recipe.Instructions, recipeInst)
		}
		recipes = append(recipes, recipe)
	}

	return recipes, nil
}

func (pg *PG) CreateRecipe(ctx context.Context, r store.Recipe) (*store.Recipe, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	tx, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating recipe: %w", err)
	}

	var recipe store.Recipe

	row := tx.QueryRowContext(ctx, sqlCreateRecipe,
		&r.UserUUID,
		&r.RecipeName,
		&r.Category,
	)

	if err = row.Scan(
		&recipe.RecipeUUID,
		&recipe.UserUUID,
		&recipe.RecipeName,
		&recipe.Category,
		&recipe.CreatedAt,
		&recipe.UpdatedAt,
	); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error creating recipe: %w", err)
	}

	for _, recipeingr := range r.Ingredients {
		var recipeIngr store.RecipeIngredient

		row := tx.QueryRowContext(ctx, sqlCreateRecipeIngr,
			&recipe.RecipeUUID, //여기 조오심..!! RecipeIngredient db model에 recipe uuid 필드가 있긴 하지만, api model에서 빈채로 요청이 오기 때문에(그리고 db에서 recipe uuid가 만들어지기 때문에) 실제로는 recipe의 recipe_uuid를 recipe_ingredients 테이블에 집어넣어야함.
			&recipeingr.IngredientUUID,
			&recipeingr.Amount,
			&recipeingr.Unit,
		)

		if err = row.Scan(
			&recipeIngr.RecipeUUID,
			&recipeIngr.IngredientUUID,
			&recipeIngr.Amount,
			&recipeIngr.Unit,
		); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("error creating recipe ingredients: %w", err)
		}
		recipe.Ingredients = append(recipe.Ingredients, recipeIngr)
	}

	for _, recipeinst := range r.Instructions {
		var recipeInst store.RecipeInstruction

		row := tx.QueryRowContext(ctx, sqlCreateRecipeInst,
			&recipe.RecipeUUID,
			&recipeinst.StepNum,
			&recipeinst.Instruction,
		)

		if err = row.Scan(
			&recipeInst.RecipeUUID,
			&recipeInst.StepNum,
			&recipeInst.Instruction,
		); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("error creating recipe instructions: %w", err)
		}
		recipe.Instructions = append(recipe.Instructions, recipeInst)
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error creating recipe: %w", err)
	}

	return &recipe, nil
}

func (pg *PG) UpdateRecipe(ctx context.Context, r store.Recipe) (*store.Recipe, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	tx, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating recipe: %w", err)
	}

	var recipe store.Recipe

	row := tx.QueryRowContext(ctx, sqlUpdateRecipe,
		r.RecipeName,
		r.Category,
		r.RecipeUUID,
	)

	if err = row.Scan(
		&recipe.RecipeUUID,
		&recipe.UserUUID,
		&recipe.RecipeName,
		&recipe.Category,
		&recipe.CreatedAt,
		&recipe.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			tx.Rollback()
			return nil, store.ErrNotFound
		}
		tx.Rollback()
		return nil, fmt.Errorf("error updating recipe: %w", err)
	}

	res, err := tx.ExecContext(ctx, sqlDeleteRecipeIngr,
		r.RecipeUUID,
	)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error deleting recipe ingredients: %w", err) //exec context는 return 값이 없을때 사용하는 거라던데, 이 update recipe는 *store.Recipe를 돌려줘야해서 nil 썼는데, 그럴거면 queryContext를 써야하나..
	}

	if _, affected := res.RowsAffected(); affected != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error deleting recipe ingredients, rows affected is %d instead of 1", affected)
	}

	for _, recipeingr := range r.Ingredients {
		var recipeIngr store.RecipeIngredient

		row := tx.QueryRowContext(ctx, sqlCreateRecipeIngr,
			&recipe.RecipeUUID,
			&recipeingr.IngredientUUID,
			&recipeingr.Amount,
			&recipeingr.Unit,
		)

		if err = row.Scan(
			&recipeIngr.RecipeUUID,
			&recipeIngr.IngredientUUID,
			&recipeIngr.Amount,
			&recipeIngr.Unit,
		); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("error creating recipe ingredients: %w", err)
		}
		recipe.Ingredients = append(recipe.Ingredients, recipeIngr)
	}

	res, err = tx.ExecContext(ctx, sqlDeleteRecipeInst,
		r.RecipeUUID,
	)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error deleting recipe instructions: %w", err)
	}

	if _, affected := res.RowsAffected(); affected != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error deleting recipe instructions, rows affected is %d instead of 1", affected)
	}

	for _, recipeinst := range r.Instructions {
		var recipeInst store.RecipeInstruction

		row := tx.QueryRowContext(ctx, sqlCreateRecipeInst,
			&recipe.RecipeUUID,
			&recipeinst.StepNum,
			&recipeinst.Instruction,
		)

		if err = row.Scan(
			&recipeInst.RecipeUUID,
			&recipeInst.StepNum,
			&recipeInst.Instruction,
		); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("error creating recipe instructions: %w", err)
		}
		recipe.Instructions = append(recipe.Instructions, recipeInst)
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error updating recipe: %w", err)
	}

	return &recipe, nil
}

func (pg *PG) DeleteRecipe(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	tx, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error deleting recipe: %w", err)
	}

	//deleting everything from recipe_ingredients table
	res, err := tx.ExecContext(ctx, sqlDeleteRecipeIngr, id)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error deleting recipe ingredients: %w", err)
	}

	if _, affected := res.RowsAffected(); affected != nil {
		tx.Rollback()
		return fmt.Errorf("error deleting recipe ingredients, rows affected is %d instead of 1", affected)
	}

	//deleting everything from recipe_instructions table
	res, err = tx.ExecContext(ctx, sqlDeleteRecipeInst, id)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error deleting recipe instructions: %w", err)
	}

	if _, affected := res.RowsAffected(); affected != nil {
		tx.Rollback()
		return fmt.Errorf("error deleting recipe instructions, rows affected is %d instead of 1", affected)
	}

	//deleting everything from recipes table
	res, err = tx.ExecContext(ctx, sqlDeleteRecipe, id)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error deleting recipe: %w", err)
	}

	if _, affected := res.RowsAffected(); affected != nil {
		tx.Rollback()
		return fmt.Errorf("error deleting recipe, rows affected is %d instead of 1", affected)
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return fmt.Errorf("error deleting recipe: %w", err)
	}

	return nil
}
