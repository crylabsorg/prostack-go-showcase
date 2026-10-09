package email

type SendWelcomeInput struct {
	ID        string `json:"ID"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"Email"`
}
