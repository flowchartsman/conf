package conf

import (
	"bufio"
	"os"
	"strings"
)

// confSource is a source for config files in an extremely simple format. Each
// line is tokenized as a single key/value pair. The first whitespace-delimited
// token in the line is interpreted as the flag name, and all remaining tokens
// are interpreted as the value. Any leading hyphens on the flag name are
// ignored.
type confSource struct {
	m map[string]string
}

func newConfSource(filename string) (*confSource, error) {
	m := make(map[string]string)

	cf, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer cf.Close()

	s := bufio.NewScanner(cf)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue // skip empties
		}

		if line[0] == '#' {
			continue // skip comments
		}

		var (
			name  string
			value string
			index = strings.IndexRune(line, ' ')
		)
		if index < 0 {
			name, value = line, "true" // boolean option
		} else {
			name, value = line[:index], strings.TrimSpace(line[index:])
		}

		if i := strings.Index(value, " #"); i >= 0 {
			value = strings.TrimSpace(value[:i])
		}

		m[name] = value
	}
	return &confSource{
		m: m,
	}, nil
}

// Get returns the stringfied value stored at the specified key in the plain
// config file
func (p *confSource) Get(key []string) (string, bool) {
	k := getEnvName(key)
	value, ok := p.m[k]
	return value, ok
}
