package configs

import (
	"os"
)

func GetEnv(envVariable string) string {
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }

	return os.Getenv(envVariable)
}
