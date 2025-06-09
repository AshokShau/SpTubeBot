package config

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
)

var (
	Token        = os.Getenv("TOKEN")
	ApiKey       = os.Getenv("API_KEY")
	ApiUrl       = os.Getenv("API_URL")
	Proxy        = os.Getenv("PROXY")
	CoolifyToken = os.Getenv("COOLIFY_TOKEN")
	DownloadPath = "downloads"
)
