package model

import "github.com/dgrijalva/jwt-go"

type MyClaims struct {
	Uuid string `json:"uuid"`
	jwt.StandardClaims
}
