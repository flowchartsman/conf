package conf

import (
	"errors"
	"fmt"
	"os"
	"reflect"
)

// ErrInvalidStruct indicates that a configuration struct is not the correct type.
var ErrInvalidStruct = errors.New("configuration must be a struct pointer")

type context struct {
	confFlag string
	confFile string
	sources  []Source
}

// Parse parses configuration into the provided struct
func Parse(confStruct interface{}, options ...Option) error {
	_, err := ParseWithArgs(confStruct, options...)
	return err
}

// ParseWithArgs parses configuration into the provided struct, returning the
// remaining args after flag parsing
func ParseWithArgs(confStruct interface{}, options ...Option) ([]string, error) {
	var c context
	for _, option := range options {
		option(&c)
	}

	fields, err := extractFields(nil, confStruct)
	if err != nil {
		return nil, err
	}

	if len(fields) == 0 {
		return nil, errors.New("no settable flags found in struct")
	}

	sources := make([]Source, 0, 3)

	// Process flags and create flag source. If help is requested, print useage
	// and exit.
	fs, args, err := newFlagSource(fields, []string{c.confFlag})
	switch err {
	case nil:
	case errHelpWanted:
		printUsage(fields, c)
		os.Exit(1)
	default:
		return nil, err
	}

	sources = append(sources, fs)

	// create config file source, if specified
	if c.confFile != "" || c.confFlag != "" {
		configFile := c.confFile
		fromFlag := false
		// if there's a config file flag, and it's set, use that filename instead
		if configFileFromFlags, ok := fs.Get([]string{c.confFlag}); ok {
			configFile = configFileFromFlags
			fromFlag = true
		}
		cs, err := newConfSource(configFile)
		if err != nil {
			if os.IsNotExist(err) {
				// The file doesn't exist. If it was specified by a flag, treat this
				// as an error, since presumably the user either made a mistake, or
				// the file they deliberately specified isn't there
				if fromFlag {
					return nil, err
				}
			} else {
				return nil, err
			}
		} else {
			sources = append(sources, cs)
		}
	}

	// create env souce
	es := new(envSource)
	sources = append(sources, es)

	// append any additional sources
	sources = append(sources, c.sources...)
	// process all fields
	if err := processFields(sources, fields); err != nil {
		// if there's an error, we should zero out all fields to avoid the case
		// where a user might not be checking the error and could end up with a
		// partially-populated struct.
		for _, f := range fields {
			f.field.Set(reflect.Zero(f.field.Type()))
		}
		return nil, err
	}

	return args, nil
}

func processFields(sources []Source, fields []field) error {
	for _, field := range fields {
		var value string
		var found bool
		for _, source := range sources {
			value, found = source.Get(field.key)
			if found {
				break
			}
		}
		if !found {
			if field.options.required {
				return fmt.Errorf("required field %s is missing value", field.name)
			}
			value = field.options.defaultStr
		}
		if value != "" {
			if err := processField(value, field.field); err != nil {
				return &processError{
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

// A processError occurs when an environment variable cannot be converted to
// the type required by a struct field during assignment.
type processError struct {
	fieldName string
	typeName  string
	value     string
	err       error
}

func (e *processError) Error() string {
	return fmt.Sprintf("conf: error assigning to field %s: converting '%s' to type %s. details: %s", e.fieldName, e.value, e.typeName, e.err)
}

// Source represents a source of configuration data. Sources requiring
// the pre-fetching and processing of several values should ideally be lazily-
// loaded so that sources further down the chain are not queried if they're
// not going to be needed.
type Source interface {
	// Get takes a location specified by a key and returns a string and whether
	// or not the value was set in the source
	Get(key []string) (value string, found bool)
}
