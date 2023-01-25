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
