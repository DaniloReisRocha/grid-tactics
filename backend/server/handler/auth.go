package handler

import (
	"net/http"
	"time"

	"github.com/DaniloReisRocha/grid-tactics/backend/internal/utils/endpoint"
	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	Middleware *jwt.GinJWTMiddleware
}

type LoginResponse struct {
	Code   int            `json:"code"`
	Expire time.Time      `json:"expire"`
	Token  string         `json:"token"`
	User   models.Account `json:"user"`
}

type RefreshTokenResponse struct {
	Code   int       `json:"code"`
	Expire time.Time `json:"expire"`
	Token  string    `json:"token"`
}

func (handler *AuthHandler) Login() endpoint.APIEndpoint {
	method := http.MethodPost
	path := "/api/login"
	return endpoint.APIEndpoint{
		Protected: false,
		NewEndpoint: func() (string, string, gin.HandlerFunc) {
			if handler.Middleware == nil {
				return method, path, func(*gin.Context) {}
			}
			return method, path, handler.Middleware.LoginHandler
		},
		MakeZod: func() (*endpoint.Zod, error) {
			return endpoint.MakeZod[models.LoginForm, endpoint.NoParams, endpoint.NoQuery, LoginResponse](method, path)
		},
	}
}

func (handler *AuthHandler) RefreshToken() endpoint.APIEndpoint {
	method := http.MethodGet
	path := "/api/refresh_token"
	return endpoint.APIEndpoint{
		Protected: false,
		NewEndpoint: func() (string, string, gin.HandlerFunc) {
			if handler.Middleware == nil {
				return method, path, func(*gin.Context) {}
			}
			return method, path, handler.Middleware.RefreshHandler
		},
		MakeZod: func() (*endpoint.Zod, error) {
			return endpoint.MakeZod[endpoint.NoBody, endpoint.NoParams, endpoint.NoQuery, RefreshTokenResponse](method, path)
		},
	}
}
