package postgres

const sqlGetUser = `
	SELECT 	user_uuid,
			hashed_password,
			active,
			first_name,
			last_name,
			email_address,
			created_at,
			updated_at
	
	FROM 	wdiet.users
	
	WHERE	user_uuid = $1

	LIMIT 1
	;
`

const sqlGetUserByEmail = `
	SELECT 	user_uuid,
			hashed_password,
			active,
			first_name,
			last_name,
			email_address,
			created_at,
			updated_at
	
	FROM 	wdiet.users
	
	WHERE	email_address = $1

	LIMIT 1
	;
`

const sqlCreateUser = `
	INSERT INTO wdiet.users(
		hashed_password,
		active,
		first_name,
		last_name,
		email_address
	)
	VALUES(
		$1,
		$2,
		$3,
		$4,
		$5
	)
	RETURNING user_uuid, hashed_password, active, first_name, last_name, email_address, created_at, updated_at
	;
`

const sqlUpdateUser = `
	UPDATE wdiet.users
		SET 
			active = $1,
			first_name = $2,
			last_name = $3,
			email_address = $4,
			updated_at = now()
	WHERE user_uuid = $5
	RETURNING user_uuid, hashed_password, active, first_name, last_name, email_address, created_at, updated_at
	;
`

const sqlGetIngredient = `
	SELECT 	ingredient_uuid,
			ingredient_name,
			category,
			days_until_exp,
			created_at,
			updated_at
	
	FROM 	wdiet.ingredients
	
	WHERE	ingredient_uuid = $1

	LIMIT 1
	;
`

const sqlsearchIngredients = `
	SELECT 	ingredient_uuid,
			ingredient_name,
			category,
			days_until_exp,
			created_at,
			updated_at

	FROM wdiet.ingredients 
`

const sqlCreateIngredient = `
	INSERT INTO wdiet.ingredients(
		ingredient_name,
		category,
		days_until_exp
	)
	VALUES(
		$1,
		$2,
		$3
	)
	RETURNING ingredient_uuid, ingredient_name, category, days_until_exp, created_at, updated_at
	;
`

const sqlUpdateIngredient = `
	UPDATE wdiet.ingredients
		SET 
			ingredient_name = $1,
			category = $2,
			days_until_exp = $3,
			updated_at = now()
	WHERE ingredient_uuid = $4
	RETURNING ingredient_uuid, ingredient_name, category, days_until_exp, created_at, updated_at
	;
`

const sqlDeleteIngredient = `
	DELETE 
		FROM wdiet.ingredients

	WHERE ingredient_uuid = $1
	;
`

const sqlListFridgeIngredients = `
	SELECT 	user_uuid,
			ingredient_uuid,
			amount,
			unit,
			purchased_date,
			expiration_date,
			created_at,
			updated_at
	
	FROM 	wdiet.fridge_ingredients
	
	WHERE	user_uuid = $1
	;
`

const sqlCreateFridgeIngredient = `
	INSERT INTO wdiet.fridge_ingredients(
				user_uuid,
				ingredient_uuid,
				amount,
				unit,
				purchased_date,
				expiration_date
	)
	VALUES(
		$1,
		$2,
		$3,
		$4,
		$5,
		$6
	)
	RETURNING user_uuid, ingredient_uuid, amount, unit, purchased_date, expiration_date, created_at, updated_at
	;
`

const sqlUpdateFridgeIngredient = `
	UPDATE wdiet.fridge_ingredients
		SET 
			amount = $1,
			unit = $2,
			purchased_date = $3,
			updated_at = now()
	WHERE user_uuid = $4 AND ingredient_uuid = $5
	RETURNING user_uuid, ingredient_uuid, amount, unit, purchased_date, expiration_date, created_at, updated_at
	;
`

const sqlDeleteFridgeIngredient = `
	DELETE 
		FROM wdiet.fridge_ingredients

	WHERE user_uuid = $1 AND ingredient_uuid = $2
	;
`
const sqlGetRecipe = `
	SELECT 	recipe_uuid,
			user_uuid,
			recipe_name,
			category,
			created_at,
			updated_at
	
	FROM 	wdiet.recipes
	
	WHERE	recipe_uuid = $1

	LIMIT 1
	;
`

// recipe 하나당 여러 recipe ingredient이므로 얘는 필연적으로 list recipe ingredients임,, 하나의 recipe ingredient만 불러오는건 의미 ㄴ
const sqlListRecipeIngr = ` 
	SELECT 	recipe_uuid,
			ingredient_uuid,
			amount,
			unit
	
	FROM 	wdiet.recipe_ingredients
	
	WHERE	recipe_uuid = $1
	;

`
const sqlListRecipeInst = `
	SELECT 	recipe_uuid,
			step_num,
			instruction
	
	FROM 	wdiet.recipe_instructions
	
	WHERE	recipe_uuid = $1
	;
`

const sqlListRecipes = `
	SELECT 	recipe_uuid,
			user_uuid,
			recipe_name,
			category,
			created_at,
			updated_at
	
	FROM 	wdiet.recipes
	
	WHERE	user_uuid = $1
	;
`

const sqlsearchRecipes = `
	SELECT 	recipe_uuid,
			user_uuid,
			recipe_name,
			category,
			created_at,
			updated_at

	FROM wdiet.recipes
`

const sqlCreateRecipe = `
	INSERT INTO wdiet.recipes(
		user_uuid,
		recipe_name,
		category
	)
	VALUES(
		$1,
		$2,
		$3
	)
	RETURNING recipe_uuid, user_uuid, recipe_name, category, created_at, updated_at
	;
`

const sqlCreateRecipeIngr = `
	INSERT INTO wdiet.recipe_ingredients(
		recipe_uuid,
		ingredient_uuid,
		amount,
		unit
	)
	VALUES(
		$1,
		$2,
		$3,
		$4
	)
	RETURNING recipe_uuid, ingredient_uuid, amount, unit
	;
`

const sqlCreateRecipeInst = `
	INSERT INTO wdiet.recipe_instructions(
		recipe_uuid,
		step_num,
		instruction
	)
	VALUES(
		$1,
		$2,
		$3
	)
	RETURNING recipe_uuid, step_num, instruction
	;
`

const sqlUpdateRecipe = `
	UPDATE wdiet.recipes
		SET 
			recipe_name = $1,
			category = $2,
			updated_at = now()
	WHERE recipe_uuid = $3
	RETURNING recipe_uuid, user_uuid, recipe_name, category, created_at, updated_at
	;
`

// AND WHERE ingredient_uuid = $2 이 부분 뺌.. 아예 특정 recipe에 해당하는 모든 ingredient들을 싹 다 지울거니까 어차피.. ingredient uuid 필요 ㄴ
// methods.go에서도 for range 안에 이걸 넣을 게 아니라 이건 바깥으로 빼놓음. 돌아가면서 ingredient들을 하나하나 지우는게 아니라 싹 다 지우고 시작.
const sqlDeleteRecipeIngr = `
	DELETE 
	FROM wdiet.recipe_ingredients

	WHERE recipe_uuid = $1
	;
`

const sqlDeleteRecipeInst = `
	DELETE 
	FROM wdiet.recipe_instructions

	WHERE recipe_uuid = $1
	;
`

const sqlDeleteRecipe = `
	DELETE 
		FROM wdiet.recipes

	WHERE recipe_uuid = $1
	;
`
