package application

import (
	"agropal/config"
	"agropal/internal/utils"
)

type WebTool struct {
	Logger *utils.Logger
	Config *config.Config
	JWT    *utils.JWT
}

func NewWebTool(
	logger *utils.Logger,
	config *config.Config,
	jwt *utils.JWT,
) WebTool {
	return WebTool{
		Logger: logger,
		Config: config,
		JWT:    jwt,
	}
}
