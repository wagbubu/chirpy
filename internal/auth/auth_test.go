package auth

import (
	"errors"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestAuth(t *testing.T) {
	dummyID := uuid.New()
	tokenSecret := "hatdog"
	t.Run("validate JWT value", func(t *testing.T) {
		token, err := MakeJWT(dummyID, tokenSecret, time.Minute)
		if err != nil {
			log.Fatal(err)
		}
		id, err := ValidateJWT(token, tokenSecret)
		if err != nil {
			log.Fatal(err)
		}
		if id != dummyID {
			t.Errorf("ID DID NOT MATCH\ngot %v\nwant %v", id, dummyID)
		}
	})
	t.Run("validate JWT expiration", func(t *testing.T) {
		token, err := MakeJWT(dummyID, tokenSecret, -time.Minute)
		if err != nil {
			log.Fatal(err)
		}
		_, err = ValidateJWT(token, tokenSecret)
		if errors.Is(err, ErrExpiredToken) {
			return
		}

		t.Errorf("token must fail validation")
	})

}
