package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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



func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn  time.Duration)(string,error){

	now := time.Now().UTC()
	claims := jwt.RegisteredClaims{
		Issuer: "chirpy",
		Subject: userID.String(),
		IssuedAt: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,claims)

	tokenString,err := token.SignedString([]byte(tokenSecret))

	if err != nil{
		return "",err
	}

	return tokenString, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error){

		claims := &jwt.RegisteredClaims{}

		token,err :=  jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any,error){

			if t.Method != jwt.SigningMethodHS256{
				return nil, errors.New("unexpected signing method")
			}
			return []byte(tokenSecret),nil

		})

		if err != nil {
			return uuid.Nil,err
		}

		if !token.Valid{
			return uuid.Nil,errors.New("invalid token")
		}

		userID, err := uuid.Parse(claims.Subject)

		return userID,nil

}
