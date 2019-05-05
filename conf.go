package conf

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidStruct indicates that a configuration struct is not the correct type.
var ErrInvalidStruct = errors.New("configuration must be a struct pointer")

// A fieldError occurs when an error occurs updating an individual field
// in the provided struct value.
type fieldError struct {
	fieldName string
	typeName  string
	value     string
	err       error
}

func (err *fieldError) Error() string {
	return fmt.Sprintf("conf: error assigning to field %s: converting '%s' to type %s. details: %s", err.fieldName, err.value, err.typeName, err.err)
}

// Source represents a source of configuration data. Sources requiring
// the pre-fetching and processing of several values should ideally be lazily-
// loaded so that sources further down the chain are not queried if they're
// not going to be needed.
type Source interface {

	// Get takes the field key and attempts to locate that key in its
	// configuration data. Returns true if found with the value.
	Get(key []string) (string, bool)
}

// Parse parses configuration into the provided struct.
func Parse(cfgStruct interface{}, sources ...Source) error {

	// Get the list of fields from the configuration struct to process.
	fields, err := extractFields(nil, cfgStruct)
	if err != nil {
		return err
	}
	if len(fields) == 0 {
		return errors.New("no fields identified in config struct")
	}

	// Process all fields found in the config struct provided.
	for _, field := range fields {
		var value string
		var found bool

		// Process each field against all sources.
		for _, source := range sources {
			value, found = source.Get(field.key)
			if found {
				break
			}
		}

		// If this key is not provided, check if required or use default.
		if !found {
			if field.options.required {
				return fmt.Errorf("required field %s is missing value", field.name)
			}
			value = field.options.defaultStr
		}

		// If this config field will be set to it's zero value, return an error.
		if value != "" {
			if err := processField(value, field.field); err != nil {
				return &fieldError{
					fieldName: field.name,
					typeName:  field.field.Type().String(),
					value:     value,
					err:       err,
				}
			}
		}
	}

	return nil
}

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
