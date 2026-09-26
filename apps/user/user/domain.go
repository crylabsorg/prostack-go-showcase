package user

import userv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/user/v1"

type User struct {
	ID    string
	Name  string
	Email string
}

func (user User) ToProto() *userv1.User {
	protoUser := &userv1.User{}
	protoUser.SetId(user.ID)
	protoUser.SetName(user.Name)
	protoUser.SetEmail(user.Email)
	return protoUser
}
