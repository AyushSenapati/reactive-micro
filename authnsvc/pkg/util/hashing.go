package util

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword takes clear text password and returns hashed password
func HashPassword(pswd string) (string, error) {
	hashedPswd, err := bcrypt.GenerateFromPassword([]byte(pswd), bcrypt.DefaultCost)
	return string(hashedPswd), err
}

// CheckPasswordHash takes clear text password and checks if its hash matches
// the given hashed password
func CheckPasswordHash(pswd, hashedPswd string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPswd), []byte(pswd))
	return err == nil
}
