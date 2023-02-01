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
	
	FROM 	wdiet.users
	
	WHERE	user_uuid = $1

	LIMIT 1
	;
`

const sqlCreateIngredient = `
	INSERT INTO wdiet.ingredients(
		ingredient_uuid,
		ingredient_name,
		category,
		days_until_exp,
	)
	VALUES(
		$1,
		$2,
		$3,
		$4,
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

	FROM ingredients

	WHERE ingredient_name LIKE '%$1%' OR category = $2 OR days_until_exp = $3
	;
`
