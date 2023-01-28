package store

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UserUUID       uuid.UUID
	HashedPassword string
	Active         bool
	FirstName      string
	LastName       string
	EmailAddress   string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Login struct {
	EmailAddress string
	Password     string
}
