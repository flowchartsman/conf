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

// Sourcer provides the ability to source data from a configuration source.
// Consider the use of lazy-loading for sourcing large datasets or systems.
type Sourcer interface {

	// Source takes the field key and attempts to locate that key in its
	// configuration data. Returns true if found with the value.
	Source(key []string) (string, bool)
}

// Parse parses configuration into the provided struct.
func Parse(cfgStruct interface{}, sources ...Sourcer) error {

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

		// Set any default value into the struct for this field.
		if field.options.defaultStr != "" {
			if err := processField(field.options.defaultStr, field.field); err != nil {
				return &fieldError{
					fieldName: field.name,
					typeName:  field.field.Type().String(),
					value:     field.options.defaultStr,
					err:       err,
				}
			}
		}

		// Process each field against all sources.
		var provided bool
		for _, sourcer := range sources {
			if sourcer == nil {
				continue
			}

			var value string
			if value, provided = sourcer.Source(field.key); !provided {
				continue
			}

			// A value was found so update the struct value with it.
			if err := processField(value, field.field); err != nil {
				return &fieldError{
					fieldName: field.name,
					typeName:  field.field.Type().String(),
					value:     value,
					err:       err,
				}
			}
		}

		// If this key is not provided by any source, check if it was
		// required to be provided.
		if !provided && field.options.required {
			return fmt.Errorf("required field %s is missing value", field.name)
		}

		// TODO : If this config field will be set to it's zero value, return an error.
		// ANDY I NEED TO UNDERSTAND WHY YOU HAD THIS. SOME PEOPLE LIKE TO BE EXPLICIT.
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
