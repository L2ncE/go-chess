package model

import "github.com/dgrijalva/jwt-go"

type MyClaims struct {
	ID   int    `json:"ID"`
	Uuid string `json:"uuid"`
	jwt.StandardClaims
}
