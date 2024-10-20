package auth

import "github.com/golang-jwt/jwt/v5"

type ContextKey string

const UserIDKey ContextKey = "uid"

type CustomClaims struct {
	UID string `json:"uid"`
	jwt.RegisteredClaims
}
