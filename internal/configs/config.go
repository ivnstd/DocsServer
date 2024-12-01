package configs

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var Config struct {
	Port       string
	AdminToken string
	Mongo      MongoConfig
}

type MongoConfig struct {
	URI  string
	Name string
}

func LoadConfig() {
	err := godotenv.Load()

	Config.Port = os.Getenv("SERVER_PORT")
	Config.AdminToken = os.Getenv("ADMIN_TOKEN")

	Config.Mongo.URI = os.Getenv("MONGO_URI")
	Config.Mongo.Name = os.Getenv("MONGO_DB_NAME")

	if err != nil {
		logrus.Fatalf("Error loading env variables: %s", err.Error())
	}
}
