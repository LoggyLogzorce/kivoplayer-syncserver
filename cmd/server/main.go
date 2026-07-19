package main

import (
	"kivo-player_sync-server/internal/models"
	"kivo-player_sync-server/internal/service"
	"log"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"kivo-player_sync-server/internal/api"
	"kivo-player_sync-server/internal/config"
	"kivo-player_sync-server/internal/handler"
	"kivo-player_sync-server/internal/repository"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file, reading from environment")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	database, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}

	if err := database.AutoMigrate(&models.Track{}, &models.Lyrics{}); err != nil {
		log.Fatalf("automigrate: %v", err)
	}

	trackRepo := repository.NewTrackRepository(database)
	trackSvc := service.NewTrackService(trackRepo)
	trackHandler := handler.NewTrackHandler(trackSvc)

	r := api.New(trackHandler)

	log.Printf("starting on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("run: %v", err)
	}
}
