package user

import (
	"github.com/crylabsorg/prostack-go-showcase/pkg/events"
	userv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/user/v1"
)

type User struct {
	ID    string
	Name  string
	Email string
}

func (user User) ToCreatedEvent() events.UserCreated {
	return events.UserCreated{ID: user.ID, Name: user.Name, Email: user.Email}
}

func (user User) ToProto() *userv1.User {
	protoUser := &userv1.User{}
	protoUser.SetId(user.ID)
	protoUser.SetName(user.Name)
	protoUser.SetEmail(user.Email)
	return protoUser
}
