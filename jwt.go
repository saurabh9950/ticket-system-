package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// A tiny hand-rolled JWT (HS256) implementation. Kept dependency-free
// on purpose so the whole service builds with zero network access.

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

type jwtClaims struct {
	Sub   string `json:"sub"`   // user ID
	Email string `json:"email"`
	Exp   int64  `json:"exp"` // unix seconds
	Iat   int64  `json:"iat"`
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

func b64url(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func b64urlDecode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

// GenerateToken creates a signed JWT for the given user, valid for ttl.
func GenerateToken(secret []byte, userID, email string, ttl time.Duration) (string, error) {
	header := jwtHeader{Alg: "HS256", Typ: "JWT"}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	claims := jwtClaims{
		Sub:   userID,
		Email: email,
		Iat:   now.Unix(),
		Exp:   now.Add(ttl).Unix(),
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	unsigned := b64url(headerJSON) + "." + b64url(claimsJSON)
	sig := signHS256(secret, unsigned)
	return unsigned + "." + b64url(sig), nil
}

// ParseToken validates the signature and expiry, returning the claims.
func ParseToken(secret []byte, token string) (*jwtClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}
	unsigned := parts[0] + "." + parts[1]
	expectedSig := signHS256(secret, unsigned)

	gotSig, err := b64urlDecode(parts[2])
	if err != nil {
		return nil, ErrInvalidToken
	}
	if subtle.ConstantTimeCompare(expectedSig, gotSig) != 1 {
		return nil, ErrInvalidToken
	}

	claimsJSON, err := b64urlDecode(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var claims jwtClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	if time.Now().UTC().Unix() > claims.Exp {
		return nil, ErrTokenExpired
	}
	return &claims, nil
}

func signHS256(secret []byte, data string) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(data))
	return mac.Sum(nil)
}
