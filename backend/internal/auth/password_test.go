package auth

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	pw := "P@ssww0d-test"
	hash, err := HashPassword(pw)

	if err != nil {
		t.Fatalf("hash error: %v", err)
	}
	if hash == pw {
		t.Fatalf("hash should not equal raw password")
	}
	if !CheckPassword(hash, pw) {
		t.Fatalf("correct password should pass")
	}
	if CheckPassword(hash, "wrong") {
		t.Fatalf("wrong password should fail")
	}
}
