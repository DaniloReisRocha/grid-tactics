package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RESTRegisterHandlers interface {
	Register(r *gin.Engine) error
}

// Config server config
type Config struct {
	Mode        string `example:"debug|release|test"`
	Version     string
	BindAddress uint16
}

// Server REST server
type Server struct {
	engine *gin.Engine
	server *http.Server
	Config Config
}

// Run start the REST server
func (srv *Server) Run(hdlr RESTRegisterHandlers) error {
	gin.SetMode(srv.Config.Mode)
	srv.engine = gin.New()

	welcome(srv.Config)
	err := hdlr.Register(srv.engine)
	if err != nil {
		return err
	}

	srv.server = &http.Server{
		Addr:    ":" + Uint16ToString(srv.Config.BindAddress),
		Handler: srv.engine,
	}

	// Start server
	return srv.server.ListenAndServe()
}

// Shutdown gracefully stops the server
func (srv *Server) Shutdown(ctx context.Context) {
	if srv.server == nil {
		log.Println("Server not started, ignoring shutdown.")
		return
	}
	if err := srv.server.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Error: %v", err)
	}
}

func NewServerWithConfig(c Config) *Server {
	return &Server{
		Config: c,
	}
}

func Uint16ToString(n uint16) string {
	return strconv.FormatUint(uint64(n), 10)
}

func welcome(config Config) {

	fmt.Printf(`   _____ _____  _____ _____    _______       _____ _______ _____ _____  _____ 
  / ____|  __ \|_   _|  __ \  |__   __|/\   / ____|__   __|_   _/ ____|/ ____|
 | |  __| |__) | | | | |  | |    | |  /  \ | |       | |    | || |    | (___  
 | | |_ |  _  /  | | | |  | |    | | / /\ \| |       | |    | || |     \___ \ 
 | |__| | | \ \ _| |_| |__| |    | |/ ____ \ |____   | |   _| || |____ ____) |
  \_____|_|  \_\_____|_____/     |_/_/    \_\_____|  |_|  |_____\_____|_____/ v%s
_________________________________________________________O/________________________
                                                         O\
⇨ HTTP server started on http://127.0.0.1:%d

`, config.Version, config.BindAddress)
}
