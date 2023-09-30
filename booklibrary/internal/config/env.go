package config

import (
	"os"
	"strconv"
)

func GetEnvString(name, value string) string {
	env, ok := os.LookupEnv(name)
	if !ok {
		return value
	}
	return env
}

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
