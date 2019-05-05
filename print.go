package conf

import (
	"fmt"
	"strings"
)

// String returns a stringified version of the provided conf-tagged
// struct, minus any fields tagged with `noprint`.
func String(v interface{}) (string, error) {
	fields, err := extractFields(nil, v)
	if err != nil {
		return "", err
	}
	var s strings.Builder
	for i, field := range fields {
		if !field.options.noprint {
			s.WriteString(field.envName)
			s.WriteString("=")
			s.WriteString(fmt.Sprintf("%v", field.field.Interface()))
			if i < len(fields)-1 {
				s.WriteString(" ")
			}
		}
	}
	return s.String(), nil
}
