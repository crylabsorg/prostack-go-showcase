package events

const UserStream = "USER_EVENTS"
const UserCreatedSubject = "user.created"

type UserCreatedEvent struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
