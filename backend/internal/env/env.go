package env

import (
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/joho/godotenv"
)

type appEnvironmentVariables struct {
	port           uint16
	allowedOrigins []string
	mode           string
}

func (app appEnvironmentVariables) Port() uint16 {
	return app.port
}

func (app appEnvironmentVariables) AllowedOrigins() []string {
	return app.allowedOrigins
}

func (app appEnvironmentVariables) Mode() string {
	return app.mode
}

var APP = appEnvironmentVariables{}

// Auth config
type authEnvironmentVariables struct {
	jwtScecret string
}

func (auth authEnvironmentVariables) JWTScecret() string {
	return auth.jwtScecret
}

var AUTH = authEnvironmentVariables{}

type databaseEnvironmentVariables struct {
	user,
	password,
	name,
	host,
	port string
}

// DB connection credentials
var DB = databaseEnvironmentVariables{}

func (d databaseEnvironmentVariables) User() string {
	return d.user
}
func (d databaseEnvironmentVariables) Password() string {
	return d.password
}
func (d databaseEnvironmentVariables) Name() string {
	return d.name
}
func (d databaseEnvironmentVariables) Host() string {
	return d.host
}
func (d databaseEnvironmentVariables) Port() string {
	return d.port
}

func Init() error {
	if len(os.Args) < 2 {
		return errors.New("Path to .env file not found")
	}

	envFile := os.Args[1]
	err := godotenv.Load(envFile)
	if err != nil {
		return errors.Wrap(err, "load .env")
	}

	// APP config
	AUTH.jwtScecret = os.Getenv("JWT_SCECRET")
	port, err := stringToUint16(os.Getenv("PORT"))
	if err != nil {
		return errors.Wrap(err, "invalid port value")
	}
	APP.port = port
	allowedOrigins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
	APP.allowedOrigins = allowedOrigins
	APP.mode = os.Getenv("MODE")

	// Database config
	DB.user = os.Getenv("DB_USER")
	DB.password = os.Getenv("DB_PASSWORD")
	DB.name = os.Getenv("DB_NAME")
	DB.host = os.Getenv("DB_HOST")
	DB.port = os.Getenv("DB_PORT")

	// Mail config

	return nil
}

func stringToUint16(s string) (uint16, error) {
	n, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(n), nil
}
