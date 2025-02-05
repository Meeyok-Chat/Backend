package configs

import (
	"os"
)

func GetEnv(envVariable string) string {
	return os.Getenv(envVariable)
}
