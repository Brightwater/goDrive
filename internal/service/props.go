package service

import (
	"log"
	"os"
	"github.com/joho/godotenv"
)

type Config struct {
    Host        string   `env:"HOST"`
    Port        string   `env:"PORT"`
    DBUser      string   `env:"DBUSER"`
    Password    string   `env:"PASSWORD"`
    DBName      string   `env:"DBNAME"`
    UploadPath  string   `env:"UPLOAD_PATH"`
    AuthBaseURL string   `env:"AUTH_BASE_PATH"`
}

var AppConfig = Config{}

func InitProps() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file %s", err)
	}

	AppConfig.AuthBaseURL = os.Getenv("AUTH_BASE_PATH")
	AppConfig.Host = os.Getenv("HOST")
	AppConfig.Port = os.Getenv("PORT")
	AppConfig.DBUser = os.Getenv("DBUSER")
	AppConfig.Password = os.Getenv("PASSWORD")
	AppConfig.DBName = os.Getenv("DBNAME")
	AppConfig.UploadPath = os.Getenv("UPLOAD_PATH")

	log.Println("Props initialized")
}