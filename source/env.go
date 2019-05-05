package source

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Env is a source for environmental variables.
type Env struct {
	m map[string]string
}

// NewEnv accepts a namespace and parses the environment into a Env for
// use by the configuration package.
func NewEnv(namespace string) (*Env, error) {
	m := make(map[string]string)

	// Get the lists of available environment variables.
	envs := os.Environ()
	if len(envs) == 0 {
		return nil, errors.New("no environment variables found")
	}

	// Create the uppercase version to meet the standard {NAMESPACE_} format.
	uspace := fmt.Sprintf("%s_", strings.ToUpper(namespace))

	// Loop and match each variable using the uppercase namespace.
	for _, val := range envs {
		if !strings.HasPrefix(val, uspace) {
			continue
		}

		idx := strings.Index(val, "=")
		m[strings.ToUpper(strings.TrimPrefix(val[0:idx], uspace))] = val[idx+1:]
	}

	// Did we find any keys for this namespace?
	if len(m) == 0 {
		return nil, fmt.Errorf("namespace %q was not found", namespace)
	}

	return &Env{m: m}, nil
}

// Get implements the confg.Source interface. It returns the stringfied value
// stored at the specified key from the environment.
func (e *Env) Get(key []string) (string, bool) {
	env := strings.ToUpper(strings.Join(key, `_`))
	return os.LookupEnv(env)
}
