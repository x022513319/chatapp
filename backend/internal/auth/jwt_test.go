package auth

import (
	"os"
	"testing"
)

func TestJWTSignAndParse(t *testing.T) {
	os.Setenv("JWT_SECRET", "test_secret")
	os.Setenv("ACCESS_TOKEN_TTL_MIN", "5")

	token, err := SignAccessToken(12345)
	if err != nil {
		t.Fatalf("sign token error: %v", err)
	}
	claims, err := ParseAndValidate(token)
	if err != nil {
		t.Fatalf("parse token error: %v", err)
	}
	if claims.Subject != "12345" {
		t.Fatalf("expected sub=12345, got %s", claims.Subject)
	}
	if claims.ID == "" { // uuid
		t.Fatalf("expected jti to be set")
	}
}
