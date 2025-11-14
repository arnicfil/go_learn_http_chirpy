package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashAndCheckPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"simple", "password123"},
		{"punctuation", "s3cure!"},
		{"unicode", "pässwørd"},
		{"empty", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			// Hash the password
			hash, err := HashPassword(tc.password)
			if err != nil {
				t.Fatalf("HashPassword returned error: %v", err)
			}
			if hash == "" {
				t.Fatalf("HashPassword returned empty hash")
			}

			// Hashing the same password twice should produce different hashes (random salt).
			hash2, err := HashPassword(tc.password)
			if err != nil {
				t.Fatalf("second HashPassword returned error: %v", err)
			}
			if hash2 == "" {
				t.Fatalf("second HashPassword returned empty hash")
			}
			if hash == hash2 {
				t.Fatalf("expected different hashes for the same password (salted), got identical")
			}

			// Correct password must validate.
			ok, err := CheckPasswordHash(tc.password, hash)
			if err != nil {
				t.Fatalf("CheckPasswordHash returned error for correct password: %v", err)
			}
			if !ok {
				t.Fatalf("CheckPasswordHash returned false for a correct password")
			}

			// Incorrect password must NOT validate (should return ok==false, err==nil).
			ok, err = CheckPasswordHash(tc.password+"x", hash)
			if err != nil {
				t.Fatalf("CheckPasswordHash returned unexpected error for incorrect password: %v", err)
			}
			if ok {
				t.Fatalf("CheckPasswordHash returned true for an incorrect password")
			}
		})
	}
}

func TestCheckPasswordHash_MalformedHash(t *testing.T) {
	t.Parallel()

	_, err := CheckPasswordHash("password", "not-a-valid-hash")
	if err == nil {
		t.Fatalf("expected an error when comparing against a malformed hash, got nil")
	}
}

func TestJWT_CreateAndValidate(t *testing.T) {
	secret := "test-secret-0123456789"
	uid := uuid.New()

	token, err := MakeJWT(uid, secret, 5*time.Minute)
	if err != nil {
		t.Fatalf("MakeJWT returned error: %v", err)
	}

	got, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT returned error for valid token: %v", err)
	}
	if got != uid {
		t.Fatalf("ValidateJWT returned wrong UUID: got %v want %v", got, uid)
	}
}

func TestJWT_ExpiredTokenIsRejected(t *testing.T) {
	t.Parallel()

	secret := "expired-secret-xyz"
	uid := uuid.New()

	token, err := MakeJWT(uid, secret, -1*time.Minute)
	if err != nil {
		t.Fatalf("MakeJWT returned error: %v", err)
	}

	_, err = ValidateJWT(token, secret)
	if err == nil {
		t.Fatalf("ValidateJWT did not return an error for an expired token")
	}
}

func TestJWT_WrongSecretIsRejected(t *testing.T) {

	secret1 := "correct-secret-abc"
	secret2 := "wrong-secret-xyz"
	uid := uuid.New()

	token, err := MakeJWT(uid, secret1, 5*time.Minute)
	if err != nil {
		t.Fatalf("MakeJWT returned error: %v", err)
	}

	_, err = ValidateJWT(token, secret2)
	if err == nil {
		t.Fatalf("ValidateJWT did not return an error when validating with the wrong secret")
	}
}

func TestJWT_MalformedTokenIsRejected(t *testing.T) {

	secret := "any-secret"
	_, err := ValidateJWT("this-is-not-a-jwt", secret)
	if err == nil {
		t.Fatalf("ValidateJWT did not return an error for a malformed token")
	}
}

func TestGetBearerToken(t *testing.T) {
	head := http.Header{}
	_, err := GetBearerToken(head)
	if err == nil {
		t.Fatalf("Error validating header without auth")
	}

	head.Add("Authorization", "something")
	_, err = GetBearerToken(head)
	if err != nil {
		t.Fatalf("Error validating header with auth")
	}
}
