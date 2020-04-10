package main

import (
	"os"
	"strings"
	"time"

	"github.com/brianvoe/sjwt"
)

var secretKey = []byte(os.Getenv("KEY"))

// VerifyToken serves to verify that the token is correct
func VerifyToken(token string) bool {
	hasVerified := sjwt.Verify(token, secretKey)
	claims, _ := sjwt.Parse(token)
	err := claims.Validate()
	if err != nil {
		return false
	}
	return hasVerified
}

// CreateToken creates the user token
func CreateToken(loginUser User, dbUser User) string {
	claims := sjwt.New()
	claims.Set("username", loginUser.Username)
	objID := dbUser.ID.String()
	x := strings.Split(objID, "(")
	id := strings.Split(x[1], ")")
	fixedID := strings.ReplaceAll(id[0], `\`, ``)
	claims.Set("id", strings.ReplaceAll(fixedID, `"`, ``))
	claims.SetExpiresAt(time.Now().Add(time.Hour * 24))
	claims.SetTokenID()
	token := claims.Generate(secretKey)
	return token
}

// GetUserIDFromToken will allow us to grab the UserID from the token
func GetUserIDFromToken(token string) string {
	if VerifyToken(token) == false {
		return "unidentified"
	}
	claims, _ := sjwt.Parse(token)
	userid, _ := claims.GetStr("id")
	return userid
}
