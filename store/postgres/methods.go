package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

	var u store.User

	row := pg.db.QueryRowContext(ctx, sqlGetUser, id)
	if err := row.Scan(
		&u.UserUUID,
		&u.HashedPassword,
		&u.Active,
		&u.FirstName,
		&u.LastName,
		&u.EmailAddress,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	return &u, nil
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

	var i store.Ingredient

	row := pg.db.QueryRowContext(ctx, sqlGetIngredient, id)
	if err := row.Scan(
		&i.IngredientUUID,
		&i.IngredientName,
		&i.Category,
		&i.DaysUntilExp,
		&i.CreatedAt,
		&i.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("error getting ingredient: %w", err)
	}

	return &i, nil
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
		&i.IngredientUUID,
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

	if affected, _ := res.RowsAffected(); affected != 1 {
		tx.Rollback()
		return fmt.Errorf("error deleting ingredient, rows affected is %d instead of 1", affected)
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return fmt.Errorf("error deleting ingredient: %w", err)
	}

	return nil
}

func (pg *PG) SearchIngredients(ctx context.Context, i store.Ingredient) ([]store.Ingredient, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var ingredients []store.Ingredient

	rows, err := pg.db.QueryContext(ctx, sqlsearchIngredients, i)
	if err != nil {
		return nil, fmt.Errorf("error searching ingredient: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var i store.Ingredient
		if err := rows.Scan(
			&i.IngredientUUID,
			&i.IngredientName,
			&i.Category,
			&i.DaysUntilExp,
		); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, store.ErrNotFound
			}
			return nil, fmt.Errorf("error searching ingredient: %w", err)
		}
		ingredients = append(ingredients, i)
	}
	// if err := rows.Close(); err != nil {
	// 	return nil, err
	// }
	// if err := rows.Err(); err != nil {
	// 	return nil, err
	// }
	return ingredients, nil
}
