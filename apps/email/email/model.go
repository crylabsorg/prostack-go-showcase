package email

import "time"

type modelWelcomeDelivery struct {
	UserID string     `bson:"_id"`
	SentAt *time.Time `bson:"sent_at,omitempty"`
}
