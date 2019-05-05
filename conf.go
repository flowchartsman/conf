package conf

import (
	"errors"
	"fmt"
	"os"
)

var (

	// ErrInvalidStruct indicates that a configuration struct is not the correct type.
	ErrInvalidStruct = errors.New("configuration must be a struct pointer")
)

type context struct {
	confFlag string // TODO: Is using conf redudant?
	confFile string
	sources  []Source
}

// Parse parses configuration into the provided struct.
func Parse(confStruct interface{}, options ...Option) error {
	_, err := ParseWithArgs(confStruct, options...)
	return err
}

// ParseWithArgs parses configuration into the provided struct, returning the
// remaining args after flag parsing.
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

	// Process flags and create a flag source.
	// If help is requested, print useage and exit.
	flagSource, args, err := newFlagSource(fields, []string{c.confFlag})
	switch err {
	case nil:
	case errHelpWanted:
		printUsage(fields, c)
		os.Exit(1)
	default:
		return nil, err
	}

	// Start collection the set of sources we need to check.
	var sources []Source
	sources = append(sources, flagSource)

	// Create the file source, if specified to do so. Then
	// add the source to the collection of sources.
	if c.confFile != "" || c.confFlag != "" {
		configFile := c.confFile
		fromFlag := false

		// If there's a config file flag, and it's set, use that filename instead.
		if configFileFromFlags, ok := flagSource.Get([]string{c.confFlag}); ok {
			configFile = configFileFromFlags
			fromFlag = true
		}

		// Create a file source for this config file.
		fileSource, err := newFileSource(configFile)
		switch {
		case err != nil:
			switch {
			case os.IsNotExist(err):

				// The file doesn't exist. If it was specified by a flag, treat this
				// as an error, since presumably the user either made a mistake, or
				// the file they deliberately specified isn't there.
				if fromFlag {
					return nil, err
				}
			default:
				return nil, err
			}
		default:
			sources = append(sources, fileSource)
		}
	}

	// Add the environment source to the list by default.
	sources = append(sources, &envSource{})

	// TODO: WERE ARE THESE COMING FROM? Append any additional source.
	sources = append(sources, c.sources...)

	// Process all fields.
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
				return nil, fmt.Errorf("required field %s is missing value", field.name)
			}
			value = field.options.defaultStr
		}
		if value != "" {
			if err := processField(value, field.field); err != nil {
				return nil, &processError{
					fieldName: field.name,
					typeName:  field.field.Type().String(),
					value:     value,
					err:       err,
				}
			}
		}
	}

	return args, nil
}

// A processError occurs when an environment variable cannot be converted to
// the type required by a struct field during assignment.
type processError struct {
	fieldName string
	typeName  string
	value     string
	err       error
}

func (pe *processError) Error() string {
	return fmt.Sprintf("conf: error assigning to field %s: converting '%s' to type %s. details: %s", pe.fieldName, pe.value, pe.typeName, pe.err)
}

// Source represents a source of configuration data. Sources requiring
// the pre-fetching and processing of several values should ideally be lazily-
// loaded so that sources further down the chain are not queried if they're
// not going to be needed.
type Source interface {

	// Get takes a location specified by a key and returns a string and whether
	// or not the value was set in the source.
	Get(key []string) (value string, found bool)
}
