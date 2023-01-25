package service

import "wdiet/store"

// func apiUser2DBUser(u User) store.User { //이거 왜써야하지? 존한테 다시 물어봐
// 	return store.User{
// 		UserUUID:     u.UserUUID,
// 		Active:       u.Active,
// 		FirstName:    u.FirstName,
// 		LastName:     u.LastName,
// 		EmailAddress: u.EmailAddress,
// 	}
// }

func dbUser2ApiUser(u store.User) User {
	return User{
		UserUUID:     u.UserUUID,
		Active:       u.Active,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		EmailAddress: u.EmailAddress,
	}
}
