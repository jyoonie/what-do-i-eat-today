package service

import "github.com/google/uuid"

func IsValidCreateUserRequest(u User, pwd string) bool {
	switch {
	case u.FirstName == "":
		return false
	case u.LastName == "":
		return false
	case u.EmailAddress == "":
		return false
	case pwd == "":
		return false
	}

	return true
}

func IsValidUpdateUserRequest(u User, uidFromPath uuid.UUID) bool {
	switch {
	case uidFromPath != u.UserUUID:
		return false
	case u.UserUUID == uuid.Nil: //when you access a field on a pointer, go has to dereference the pointer. If you pass in a nil there, you'll panic, because you're dereferencing a nil pointer.(that's why you want to return *struct from a db method too.)
		return false
	case u.FirstName == "":
		return false
	case u.LastName == "":
		return false
	case u.EmailAddress == "":
		return false
	}

	return true
}
