package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
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

func GetBearerToken(headers http.Header) (string, error) {
    authHeader := headers.Get("Authorization")
    if authHeader == "" {
        return "", errors.New("authorization header missing")
    }

    const prefix = "Bearer "
    if !strings.HasPrefix(authHeader, prefix) {
        return "", errors.New("authorization header must start with 'Bearer '")
    }

    token := strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
    if token == "" {
        return "", errors.New("token is empty")
    }

    return token, nil
}

func MakeRefreshToken() (string, error){

	bytes := make([]byte, 32)

	_,err := rand.Read(bytes)
	if err != nil{
		return "",err
	}
	token := hex.EncodeToString(bytes)

	return token,nil
}

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header missing")
	}

	const prefix = "ApiKey "
	if !strings.HasPrefix(authHeader, prefix) {
		return "", errors.New("authorization header must start with 'ApiKey '")
	}

	apiKey := strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
	if apiKey == "" {
		return "", errors.New("API key is empty")
	}

	return apiKey, nil
}
