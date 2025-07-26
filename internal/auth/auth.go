package auth

import (
	"golang.org/x/crypto/bcrypt"
)


func HashPassword(password string) (string, error){

	hasedPassword,err := bcrypt.GenerateFromPassword([]byte(password),bcrypt.DefaultCost)

	if err != nil {
		return "Error hashing password",err
	}

	return string(hasedPassword),nil
	
}

func CheckPasswordHash(password, hash string) error{
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}