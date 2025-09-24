package env

import (
	"os"
	"strconv"
)

// get env value of key or default
func GetString(key, def string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	return val
}

func GetInt(key string, def int) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return def
	}
	return i
}

func GetNonEmptyString(key, def string) string {
	val := GetString(key, def)
	if val == "" {
		return def
	}
	return val
}
