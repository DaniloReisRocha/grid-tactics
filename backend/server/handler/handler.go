package handler

import (
	"log"
	"net/http"
	"os"

	"github.com/DaniloReisRocha/grid-tactics/backend/internal/utils/endpoint"
	"github.com/DaniloReisRocha/grid-tactics/backend/server/middleware"
	"github.com/DaniloReisRocha/grid-tactics/backend/service"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type ContextQ struct {
	Context string `json:"context,omitempty"`
}

type JWTConfig struct {
	Secret,
	ClaimsCtxKey string
}

type ServerConf struct {
	BodyLimit      string
	AllowedOrigins []string
}

type HTTPServer struct {
	DB         *sqlx.DB
	ServerConf ServerConf
	JWTConfig  JWTConfig
}

func (h *HTTPServer) validateDeps() error {
	if h.DB == nil {
		return errors.New("missing dependency")
	}
	if len(h.ServerConf.AllowedOrigins) == 0 {
		return errors.New("all origins disabled")
	}
	return nil
}

func (h *HTTPServer) Register(engine *gin.Engine) error {
	err := h.validateDeps()
	if err != nil {
		return err
	}

	engine.Use(gin.Recovery())
	engine.Use(gin.Logger())

	// CORS
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     h.ServerConf.AllowedOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	// Session middleware
	store := cookie.NewStore([]byte(h.JWTConfig.Secret))
	engine.Use(sessions.Sessions("mysession", store))

	// JWT middleware
	authenticator := service.AuthenticatorService{DB: h.DB}
	authMiddleware, err := middleware.NewJWTMiddleware(h.JWTConfig.Secret, authenticator.Run)
	if err != nil {
		log.Printf("JWT Middleware init failed: %v", err)
		return err
	}

	// Protected group
	protected := engine.Group("")
	protected.Use(authMiddleware.MiddlewareFunc())
	protected.Use(middleware.RateLimitWithConfig(middleware.RateLimitConfig{
		Limit: 10,
		Burst: 20,
	}))

	for _, ep := range h.Routes(authMiddleware) {
		if ep.Protected {
			protected.Handle(ep.NewEndpoint())
		} else {
			engine.Handle(ep.NewEndpoint())
		}
	}

	return nil
}

func (h *HTTPServer) Routes(authMiddleware *jwt.GinJWTMiddleware) []endpoint.APIEndpoint {
	// auth
	authH := AuthHandler{Middleware: authMiddleware}

	gsG := service.GameUnitsGetService{DB: h.DB}
	gsWs := service.GameWsService{
		AllowedOrigins: h.ServerConf.AllowedOrigins,
		GetUnits:       gsG.Run,
		Logger:         log.New(os.Stderr, "PlaygroundWs: ", log.Lshortfile),
	}
	gsH := GameSocketHandler{
		gameWs: gsWs.Receive,
	}

	// Example
	exC := service.ExampleCreateService{DB: h.DB}
	exG := service.ExampleGetService{DB: h.DB}
	exL := service.ExampleListService{DB: h.DB}
	exU := service.ExampleUpdateService{DB: h.DB}
	exD := service.ExampleDeleteService{DB: h.DB}
	exH := ExampleHandler{
		create: exC.Run,
		get:    exG.Run,
		list:   exL.Run,
		update: exU.Run,
		delete: exD.Run,
	}

	return []endpoint.APIEndpoint{
		// auth routes
		authH.Login(),
		authH.RefreshToken(),

		// Game soket
		gsH.GameWs(),

		// Example routes
		exH.Create(),
		exH.Update(),
		exH.Delete(),
		exH.Get(),
		exH.List(),
	}
}
