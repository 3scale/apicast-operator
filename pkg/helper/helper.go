package helper

import "os"

func GetEnvVar(key, def string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return def
}

func Int32Ptr(i int32) *int32 {
	return &i
}
