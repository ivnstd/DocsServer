package main

import (
	"context"
	"docs_server/internal/configs"
	"docs_server/internal/handler"
	"docs_server/internal/repository"
	"docs_server/internal/service"
	"docs_server/pkg/db"
	"docs_server/pkg/server"

	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	configs.LoadConfig()
	cfg := configs.Config

	logrus.Info("Starting server...")

	db, err := db.NewMongoDB(cfg.Mongo.URI, cfg.Mongo.Name)
	if err != nil {
		logrus.Fatalf("Failed to initialize db: %s", err.Error())
	}
	logrus.Info("Database connection established")

	repos := repository.NewRepository(db)
	services := service.NewService(repos)
	handlers := handler.NewHandler(services)
	router := handlers.InitRoutes()

	go repos.Users.RemoveExpiredSessions(context.Background())

	srv := new(server.Server)
	if err := srv.Run(cfg.Port, router); err != nil {
		logrus.Fatalf("Error occured while running http server: %s", err.Error())
	}
	logrus.Info("HTTP server successfully launched")
}
