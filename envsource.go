package conf

import "os"

type envSource struct{}

func (e *envSource) Get(key []string) (string, bool) {
	varName := getEnvName(key)
	return os.LookupEnv(varName)
}
