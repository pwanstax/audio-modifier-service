package services

import (
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type JWTService interface {
	GenerateToken() string
	ValidateToken(tokenString string) (*jwt.Token, error)
}

type jwtCustomClaims struct {
	jwt.StandardClaims
}

type jwtService struct {
	secretKey string
	issuer    string
}

func NewJWTService() JWTService {
	return &jwtService{
		secretKey: getSecretKey(),
		issuer:    "Depression",
	}
}

func getSecretKey() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "Uk1(I91icOLA01lIVK,SSiL!_:C.foD0ike"
	}
	return secret

}

func (jwtSrv *jwtService) GenerateToken() string {
	claims := &jwtCustomClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
			Issuer:    jwtSrv.issuer,
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(jwtSrv.secretKey))
	if err != nil {
		panic(err)
	}
	return t

}

func (jwtSrv *jwtService) ValidateToken(tokenString string) (*jwt.Token, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("authorization cannot be empty")
	}
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])

		}
		return []byte(jwtSrv.secretKey), nil

	})
}
