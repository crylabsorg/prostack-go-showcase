package email

import "github.com/crylabsorg/prostack-go-showcase/pkg/events"

type SendWelcomeInput struct {
	ID        string `json:"ID"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"Email"`
}

func newSendWelcomeInput(event events.UserCreatedEvent) SendWelcomeInput {
	input := SendWelcomeInput{ID: event.ID, Email: event.Email}
	if event.FirstName != nil {
		input.FirstName = *event.FirstName
	}
	if event.LastName != nil {
		input.LastName = *event.LastName
	}
	return input
}
