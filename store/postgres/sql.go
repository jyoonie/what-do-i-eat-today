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
