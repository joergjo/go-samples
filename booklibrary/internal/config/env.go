package config

import (
	"os"
	"strconv"
)

// GetEnvString returns the value of the environment variable named by the key, or the default value if the environment variable is not set.
func GetEnvString(name, value string) string {
	env, ok := os.LookupEnv(name)
	if !ok {
		return value
	}
	return env
}

// GetEnvInt returns the value of the environment variable named by the key, or the default value if the environment variable is not set or is not a valid integer.
func GetEnvInt(name string, value int) int {
	envStr, ok := os.LookupEnv(name)
	if !ok {
		return value
	}
	env, err := strconv.Atoi(envStr)
	if err != nil {
		return value
	}
	return env
}

// GetEnvBool returns the value of the environment variable named by the key, or the default value if the environment variable is not set or is not a valid bool.
func GetEnvBool(name string, value bool) bool {
	envStr, ok := os.LookupEnv(name)
	if !ok {
		return value
	}
	env, err := strconv.ParseBool(envStr)
	if err != nil {
		return value
	}
	return env
}
