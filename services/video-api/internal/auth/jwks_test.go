package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestValidatorAcceptsValidTokenAndCachesJWKS(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	server := jwksServer(t, &privateKey.PublicKey, "key-1")
	defer server.Close()

	validator := New("issuer", "audience", server.URL)
	token := signedToken(t, privateKey, "key-1", "issuer", "audience", "subject-1")
	subject, err := validator.Validate(context.Background(), token)
	if err != nil || subject != "subject-1" {
		t.Fatalf("subject=%q err=%v", subject, err)
	}

	server.Close()
	subject, err = validator.Validate(context.Background(), token)
	if err != nil || subject != "subject-1" {
		t.Fatalf("cache subject=%q err=%v", subject, err)
	}
}

func TestValidatorRejectsInvalidClaimsAndUnknownKey(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	server := jwksServer(t, &privateKey.PublicKey, "key-1")
	defer server.Close()
	validator := New("issuer", "audience", server.URL)

	if _, err := validator.Validate(context.Background(), signedToken(t, privateKey, "other", "issuer", "audience", "sub")); err == nil {
		t.Fatal("expected unknown key error")
	}
	if _, err := validator.Validate(context.Background(), signedToken(t, privateKey, "key-1", "wrong", "audience", "sub")); err == nil {
		t.Fatal("expected issuer error")
	}
}

func jwksServer(t *testing.T, key *rsa.PublicKey, kid string) *httptest.Server {
	t.Helper()
	n := base64.RawURLEncoding.EncodeToString(key.N.Bytes())
	e := big.NewInt(int64(key.E)).Bytes()
	exponent := base64.RawURLEncoding.EncodeToString(e)
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{"kid": kid, "kty": "RSA", "alg": "RS256", "n": n, "e": exponent}}})
	}))
}

func signedToken(t *testing.T, key *rsa.PrivateKey, kid, issuer, audience, subject string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": issuer, "aud": audience, "sub": subject,
		"exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix(),
	})
	token.Header["kid"] = kid
	raw, err := token.SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
