package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	if os.Getenv("APP_ENV") == "local" {
		godotenv.Load("local.env")
	}
}

func GetEnv(env string) string {
	return os.Getenv(env)
}

func GetEnvInt(env string) int {
	value := os.Getenv(env)
	intValue, err := strconv.Atoi(value)
	if err != nil {
		fmt.Printf("Environment variable %s is not a valid integer: %s", env, value)
	}
	return intValue
}

func GetEnvBool(env string) bool {
	value := os.Getenv(env)
	intValue, err := strconv.ParseBool(value)
	if err != nil {
		fmt.Printf("Environment variable %s is not a valid bool: %s", env, value)
	}
	return intValue
}
