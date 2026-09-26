package user

type modelUser struct {
	ID    string
	Name  string
	Email string
}

func (u modelUser) domain() User {
	return User{ID: u.ID, Name: u.Name, Email: u.Email}
}
