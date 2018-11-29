package authentication

import (
	jwt "github.com/dgrijalva/jwt-go"
)

type User struct {
	UUID       string `json:"uuid" form:"-"`
	Username   string `json:"username" form:"username"`
	Password   string `json:"password" form:"password"`
	Repository string `json:"repository"`
}

type Authentifier interface {
	GenerateToken(userUUID string) (string, error)
	Authenticate(user, storedUser User) bool
	Logout(tokenString string, token *jwt.Token) error
	Login(requestUser, dbUser *User) (int, []byte)
}
