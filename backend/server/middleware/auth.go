package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

const identityKey = "accoID"

func NewJWTMiddleware(secret string, authenticator models.Authenticator) (*jwt.GinJWTMiddleware, error) {
	return jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "LEASE MANAGER API",
		Key:         []byte(secret),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour,
		IdentityKey: identityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if user, ok := data.(*models.Account); ok {
				return jwt.MapClaims{
					identityKey: user.AccoID,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			id := claims[identityKey].(string)
			accoID, err := uuid.FromString(id)
			if err != nil {
				return nil
			}
			return accoID
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginForm models.LoginForm
			if err := c.ShouldBindJSON(&loginForm); err != nil {
				return "", jwt.ErrMissingLoginValues
			}
			data, err := authenticator(loginForm)
			if err != nil {
				return nil, err
			}
			c.Set("user", data)
			return data, nil
		},
		LoginResponse: func(c *gin.Context, code int, token string, expire time.Time) {
			g := gin.H{"token": token, "expire": expire}
			if user, exists := c.Get("user"); exists {
				g["code"] = code
				g["user"] = user
				c.JSON(code, g)
				return
			}
			c.JSON(code, g)
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			_, ok := data.(uuid.UUID)
			return ok
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			fmt.Println("Unauthorized Message: ", message)
			c.JSON(code, gin.H{"error": message})
		},
		TokenLookup:   "header: Authorization, query: token, cookie: jwt",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
	})
}

func ValidateAccoID(c *gin.Context) (uuid.UUID, bool) {
	identity, exists := c.Get(identityKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing identity"})
		c.Abort()
		return uuid.UUID{}, false
	}

	accoID, ok := identity.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: invalid account identifier"})
		c.Abort()
		return uuid.UUID{}, false
	}

	return accoID, true
}
