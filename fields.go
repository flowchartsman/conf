package conf

import (
	"reflect"
)

// field maintains information about a field in the configuration struct
type field struct {
	name    string
	key     []string
	field   reflect.Value
	options fieldOptions
	// important for flag parsing or any other source where booleans might be
	// treated specially
	boolField bool
	// for usage
	flagName string
	envName  string
}

type fieldOptions struct {
	//allow for alternate name, perhaps
	short      rune
	help       string
	defaultStr string
	noprint    bool
	required   bool
}

// extractFields uses reflection to examine the struct and generate the keys
func extractFields(prefix []string, target interface{}, c context) ([]field, error) {
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

	fields := []field{}

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		structField := targetType.Field(i)

		// get the conf tags associated with this item (if any)
		fieldTags := structField.Tag.Get("conf")

		// if it's ignored or can't be set, move on
		if !f.CanSet() || fieldTags == "-" {
			continue
		}

		fieldOpts, err := parseTag(fieldTags)
		if err != nil {
			return nil, err
		}

		// found a pointer
		for f.Kind() == reflect.Ptr {
			if f.IsNil() {
				// if it's not a struct, we don't care
				if f.Type().Elem().Kind() != reflect.Struct {
					break
				}
				// if it is, create a zero instance
				f.Set(reflect.New(f.Type().Elem()))
			}
			f = f.Elem()
		}

		fieldName := structField.Name
		// break name into constituent pieces via CamelCase parser
		fieldKey := append(prefix, camelSplit(fieldName)...)

		// if we've found a struct, drill down, appending fields as we go
		if f.Kind() == reflect.Struct {
			// skip if it can deserialize itself
			if setterFrom(f) == nil && textUnmarshaler(f) == nil && binaryUnmarshaler(f) == nil {
				// prefix for any subkeys is the fieldKey, unless it's anonymous, then it's just the prefix so far
				innerPrefix := fieldKey
				if structField.Anonymous {
					innerPrefix = prefix
				}

				embeddedPtr := f.Addr().Interface()
				innerFields, err := extractFields(innerPrefix, embeddedPtr, c)
				if err != nil {
					return nil, err
				}
				fields = append(fields, innerFields...)
			}
		} else {
			// append the field
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
