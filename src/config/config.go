package config

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
)

var (
	ApiId        = os.Getenv("API_ID")
	ApiHash      = os.Getenv("API_HASH")
	Token        = os.Getenv("TOKEN")
	ApiKey       = os.Getenv("API_KEY")
	ApiUrl       = os.Getenv("API_URL")
	CoolifyToken = os.Getenv("COOLIFY_TOKEN")
	DownloadPath = "downloads"
)
