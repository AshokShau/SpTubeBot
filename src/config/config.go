package config

import (
	_ "github.com/joho/godotenv/autoload"
	"os"
)

var (
	Token        = os.Getenv("TOKEN")
	ApiKey       = os.Getenv("API_KEY")
	ApiUrl       = os.Getenv("API_URL")
	CoolifyToken = os.Getenv("COOLIFY_TOKEN")
	DownloadPath = "downloads"
)
