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
		tx.Rollback()
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	return &user, nil
}
