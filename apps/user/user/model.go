package user

type modelUser struct {
	ID        string
	FirstName *string
	LastName  *string
	Email     string
}

func (u modelUser) toDomain() User {
	return User(u)
}
