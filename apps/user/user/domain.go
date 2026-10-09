package user

import (
	"github.com/crylabsorg/prostack-go-showcase/pkg/events"
	userv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/user/v1"
)

type User struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
}

func (user User) ToCreatedEvent() events.UserCreatedEvent {
	return events.UserCreatedEvent{ID: user.ID, FirstName: user.FirstName, LastName: user.LastName, Email: user.Email}
}

func (user User) ToProto() *userv1.User {
	protoUser := &userv1.User{}
	protoUser.SetId(user.ID)
	protoUser.SetFirstName(user.FirstName)
	protoUser.SetLastName(user.LastName)
	protoUser.SetEmail(user.Email)
	return protoUser
}

type OutboxMessage struct {
	id      string
	subject string
	payload []byte
}

func (message OutboxMessage) ID() string { return message.id }

func (message OutboxMessage) Subject() string { return message.subject }

func (message OutboxMessage) Payload() []byte { return message.payload }

type OutboxPublishResult struct {
	ID  string
	Err error
}
