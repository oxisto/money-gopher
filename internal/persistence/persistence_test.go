package persistence

import (
	"context"
	"testing"
)

func TestOpenDB(t *testing.T) {
	db, err := OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB() error = %v", err)
	}
	defer db.Close()

	user, err := db.CreateUser(context.Background(), CreateUserParams{
		ID:          "user-1",
		Issuer:      "https://issuer.example.com",
		Subject:     "subject-1",
		DisplayName: "Money Gopher",
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	got, err := db.GetUserByIdentity(context.Background(), GetUserByIdentityParams{
		Issuer:  "https://issuer.example.com",
		Subject: "subject-1",
	})
	if err != nil {
		t.Fatalf("GetUserByIdentity() error = %v", err)
	}
	if got.ID != user.ID {
		t.Errorf("GetUserByIdentity() ID = %v, want %v", got.ID, user.ID)
	}
}
