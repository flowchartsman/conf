package conf

import (
	"fmt"
	"reflect"
	"strings"
)

// field maintains information about a field in the configuration struct.
type field struct {
	name    string
	key     []string
	field   reflect.Value
	options fieldOptions

	// Important for flag parsing or any other source where booleans might be
	// treated specially.
	boolField bool

	// For usage ...  TODO: I need more.
	flagName string
	envName  string
}

type fieldOptions struct {
	short      rune // Allow for alternate name, perhaps.
	help       string
	defaultStr string
	noprint    bool
	required   bool
}

// extractFields uses reflection to examine the struct and generate the keys.
func extractFields(prefix []string, target interface{}) ([]field, error) {
	if prefix == nil {
		prefix = []string{}
	}
	s := reflect.ValueOf(target)

	if s.Kind() != reflect.Ptr {
		return nil, ErrInvalidStruct
	}
	s = s.Elem()
	if s.Kind() != reflect.Struct {
		return nil, ErrInvalidStruct
	}
	targetType := s.Type()

	var fields []field

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		structField := targetType.Field(i)

		// Get the conf tags associated with this item (if any).
		fieldTags := structField.Tag.Get("conf")

		// If it's ignored or can't be set, move on.
		if !f.CanSet() || fieldTags == "-" {
			continue
		}

		fieldName := structField.Name

		// Break name into constituent pieces via CamelCase parser.
		fieldKey := append(prefix, camelSplit(fieldName)...)

		// Get and options.  TODO: Need more.
		fieldOpts, err := parseTag(fieldTags)
		if err != nil {
			return nil, fmt.Errorf("conf: error parsing tags for field %s: %s", fieldName, err)
		}

		// Drill down through pointers until we bottom out at type or nil.
		for f.Kind() == reflect.Ptr {
			if f.IsNil() {

				// It's not a struct so leave it alone.
				if f.Type().Elem().Kind() != reflect.Struct {
					break
				}

				// It is a struct so zero it out.
				f.Set(reflect.New(f.Type().Elem()))
			}
			f = f.Elem()
		}

		switch {

		// If we've found a struct, drill down, appending fields as we go.
		case f.Kind() == reflect.Struct:

			// Skip if it can deserialize itself.
			if setterFrom(f) == nil && textUnmarshaler(f) == nil && binaryUnmarshaler(f) == nil {

				// Prefix for any subkeys is the fieldKey, unless it's
				// anonymous, then it's just the prefix so far.
				innerPrefix := fieldKey
				if structField.Anonymous {
					innerPrefix = prefix
				}

				embeddedPtr := f.Addr().Interface()
				innerFields, err := extractFields(innerPrefix, embeddedPtr)
				if err != nil {
					return nil, err
				}
				fields = append(fields, innerFields...)
			}
		default:
			fields = append(fields, field{
				name:      fieldName,
				key:       fieldKey,
				flagName:  getFlagName(fieldKey),
				envName:   getEnvName(fieldKey),
				field:     f,
				options:   fieldOpts,
				boolField: f.Kind() == reflect.Bool,
			})
		}
	}

	return fields, nil
}

func parseTag(tagStr string) (fieldOptions, error) {
	var f fieldOptions
	if tagStr == "" {
		return f, nil
	}

	tagParts := strings.Split(tagStr, ",")
	for _, tagPart := range tagParts {
		vals := strings.SplitN(tagPart, ":", 2)
		tagProp := vals[0]

		switch len(vals) {
		case 1:
			switch tagProp {
			case "noprint":
				f.noprint = true
			case "required":
				f.required = true
			}
		case 2:
			tagPropVal := strings.TrimSpace(vals[1])
			if tagPropVal == "" {
				return f, fmt.Errorf("tag %q missing a value", tagProp)
			}
			switch tagProp {
			case "short":
				if len([]rune(tagPropVal)) != 1 {
					return f, fmt.Errorf("short value must be a single rune, got %q", tagPropVal)
				}
				f.short = []rune(tagPropVal)[0]
			case "default":
				f.defaultStr = tagPropVal
			case "help":
				f.help = tagPropVal
			}
		default:
			// TODO: Do we check for integrity issues here?
		}
	}

	// Perform a sanity check.
	switch {
	case f.required && f.defaultStr != "":
		return f, fmt.Errorf("cannot set both `required` and `default`")
	}

	return f, nil
}
