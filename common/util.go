package common

import "os"

func Env(key, def string) string {
	var val string
	val = os.Getenv(key)
	val, set := os.LookupEnv(key)
	if set {
		return val
	} else {
		return def
	}
}

func BoolEnv(key, def string) bool {
	val := Env(key, def)

	return val == "true"
}
