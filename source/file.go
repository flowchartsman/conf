package source

import (
	"bufio"
	"os"
	"strings"
)

// File is a source for config files in an extremely simple format. Each
// line is tokenized as a single key/value pair. The first whitespace-delimited
// token in the line is interpreted as the flag name, and all remaining tokens
// are interpreted as the value. Any leading hyphens on the flag name are
// ignored.
type File struct {
	m map[string]string
}

// NewFile accepts a filename and parses the contents into a File for
// use by the configuration package.
func NewFile(filename string) (*File, error) {
	m := make(map[string]string)

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	s := bufio.NewScanner(file)
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

	return &File{m: m}, nil
}

// Source implements the confg.Sourcer interface. It returns the stringfied value
// stored at the specified key in the plain config file.
func (f *File) Source(key []string) (string, bool) {
	k := strings.ToUpper(strings.Join(key, `_`))
	value, ok := f.m[k]
	return value, ok
}
