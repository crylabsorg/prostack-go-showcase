package events

const UserStream = "USER_EVENTS"
const UserCreatedSubject = "user.created"

type UserCreatedEvent struct {
	ID        string  `json:"id"`
	FirstName *string `json:"first_name,omitzero"`
	LastName  *string `json:"last_name,omitzero"`
	Email     string  `json:"email"`
}
