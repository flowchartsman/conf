package conf

import "os"

type envSource struct{}

// Get returns the stringfied value stored at the specified key
// from the environment.
func (e *envSource) Get(key []string) (string, bool) {
	varName := getEnvName(key)
	return os.LookupEnv(varName)
}
