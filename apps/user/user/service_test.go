package user

import (
	"context"
	"errors"
	"testing"
)

func TestServiceRejectsInvalidInput(t *testing.T) {
	svc := &Service{}
	if _, err := svc.Create(context.Background(), "Ada", "not-an-email"); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Create invalid email: got %v", err)
	}
	if _, err := svc.Get(context.Background(), "not-a-uuid"); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Get invalid ID: got %v", err)
	}
}
