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
