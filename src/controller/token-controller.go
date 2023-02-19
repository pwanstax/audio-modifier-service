package controller

import (
	"server/services"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type TokenController interface {
	GenerateToken() string
	ValidateToken(ctx *gin.Context) (*jwt.Token, error)
}

type tokenController struct {
	jwtService services.JWTService
}

func NewTokenController() TokenController {
	return &tokenController{
		jwtService: services.NewJWTService(),
	}
}

func (tokenCtl *tokenController) GenerateToken() string {
	return tokenCtl.jwtService.GenerateToken()

}

func (tokenCtl *tokenController) ValidateToken(ctx *gin.Context) (*jwt.Token, error) {
	token := ctx.Request.Header.Get("token")
	return tokenCtl.jwtService.ValidateToken(token)

}
