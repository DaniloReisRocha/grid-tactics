package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/DaniloReisRocha/grid-tactics/backend/internal/env"
	"github.com/DaniloReisRocha/grid-tactics/backend/server"
	"github.com/DaniloReisRocha/grid-tactics/backend/server/handler"
	_ "github.com/lib/pq"

	"github.com/jmoiron/sqlx"
)

func main() {
	if err := env.Init(); err != nil {
		log.Fatalf("Failed to load envs: %v", err)
	}

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		env.DB.Host(), env.DB.Port(), env.DB.User(), env.DB.Password(), env.DB.Name())

	db, err := sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(100)

	serv := server.NewServerWithConfig(server.Config{
		BindAddress: env.APP.Port(),
		Version:     "0.1",
		Mode:        env.APP.Mode(),
	})

	handler := handler.HTTPServer{
		DB: db,
		JWTConfig: handler.JWTConfig{
			Secret: env.AUTH.JWTScecret(),
		},
		ServerConf: handler.ServerConf{
			BodyLimit:      "",
			AllowedOrigins: env.APP.AllowedOrigins(),
		},
	}

	errChan := make(chan error, 1)
	go func() {
		if err := serv.Run(&handler); err != nil {
			errChan <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	select {
	case <-quit:
		log.Println("Received interrupt signal. Shutting down server...")
	case err := <-errChan:
		log.Fatalf("Failed to start server: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	serv.Shutdown(ctx)
}
