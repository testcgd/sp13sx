package util

import "os"

func LookupEnv(name string) (string, bool) {
	if name == "" {
		return "", false
	}
	return os.LookupEnv(name)
}
