package service

import "wdiet/store"

func apiUser2DBUser(u User) store.User { //your user always submits a value, that's why you take the value. 이 모델은 핸들러에서 요청 받은거 db model로 변환할 때 쓴다.
	return store.User{
		UserUUID:     u.UserUUID, //for every user related request, this field is needed, but for createuser request, you don't need this fied. So you can just leave it empty.
		Active:       u.Active,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		EmailAddress: u.EmailAddress,
	}
}

func dbUser2ApiUser(u *store.User) User { //your db always spits out the pointer, that's why you take the pointer.
	return User{
		UserUUID:     u.UserUUID,
		Active:       u.Active,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		EmailAddress: u.EmailAddress,
	}
}
