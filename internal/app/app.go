package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/BabyJhon/medods-test-task/internal/handlers"
	"github.com/BabyJhon/medods-test-task/internal/repo"
	"github.com/BabyJhon/medods-test-task/internal/service"
	"github.com/BabyJhon/medods-test-task/pkg/httpserver"
	"github.com/BabyJhon/medods-test-task/pkg/postgres"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func Run() {
	logrus.SetFormatter(new(logrus.JSONFormatter))

	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("error loading env vars: %s", err.Error())
	}

	pool, err := postgres.NewPG(context.Background(), postgres.Config{
		Host:     os.Getenv("PG_HOST"),
		Port:     os.Getenv("PG_PORT"),
		Username: os.Getenv("PG_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("PG_DATABASE_NAME"),
		SSLMode:  os.Getenv("PG_SSLMODE"),
	})
	if err != nil {
		logrus.Fatalf("failed init db: %s", err.Error())
	}

	defer pool.Close()

	repos := repo.NewRepository(pool)

	services := service.NewService(repos)

	handlers := handlers.NewHandler(services)

	srv := new(httpserver.Server)

	go func() {
		if err := srv.Run(os.Getenv("PORT"), handlers.InitRoutes()); err != http.ErrServerClosed {
			logrus.Fatalf("error occured while running server: %s", err.Error())
		}
	}()

	logrus.Print("API started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logrus.Print("shutting down")
	if err := srv.ShutDown(context.Background()); err != nil {
		logrus.Errorf("error while server shutting down: %s", err.Error())
	}

}
