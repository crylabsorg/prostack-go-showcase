package user

type modelUser struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
}

func (u modelUser) toDomain() User {
	return User{ID: u.ID, FirstName: u.FirstName, LastName: u.LastName, Email: u.Email}
}
