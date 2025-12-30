package fx

import (
	"log"

	"Fynance/config"
	"Fynance/internal/logger"

	"github.com/joho/godotenv"
	"go.uber.org/fx"
)

var ConfigModule = fx.Module("config",
	fx.Provide(
		config.Load,
	),
	fx.Invoke(
		loadEnvFiles,
		initLogger,
	),
)

func loadEnvFiles() error {
	if err := godotenv.Load(); err != nil {
		log.Printf("Aviso: não foi possível carregar .env do diretório atual: %v", err)
	}
	if err := godotenv.Load("../../.env"); err != nil {
		log.Printf("Aviso: não foi possível carregar ../../.env: %v", err)
	}
	return nil
}

func initLogger(cfg *config.Config) {
	logger.Init(cfg)
}
