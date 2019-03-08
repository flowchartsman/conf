package conf

import (
	"fmt"
	"os"
	"reflect"
	"sort"
)

func printUsage(fields []field, c context) {
	if c.confFlag != "" {
		confFlagField := field{
			flagName: c.confFlag,
		}
		if c.confFile != "" {
			confFlagField.options.defaultStr = c.confFile
			confFlagField.options.help = "the 'filename' to load configuration from"
		}
		fields = append(fields, confFlagField)
	}

	// sort the fields, by their long name
	sort.SliceStable(fields, func(i, j int) bool {
		return fields[i].flagName < fields[j].flagName
	})

	uprint("Usage: %s [options] [arguments]\n\nOPTIONS\n", os.Args[0])
	for _, f := range fields {
		name, help := unquoteHelp(&f)
		uprint("\t")
		uprint("--%s", f.flagName)
		if f.options.short != 0 {
			uprint(", -%s", string(f.options.short))
		}
		if f.boolField {
			if help != "" {
				uprint("\t%s", help)
			}
		} else {
			uprint(" %s", name)
			if help != "" {
				uprint("\n\t\t%s", help)
			}
		}
		if f.options.defaultStr != "" {
			uprint("\n\t\t(default: %s)", f.options.defaultStr)
		}
		uprint("\n")
	}
	uprint("\n")
	if c.confFile != "" {
		uprint("FILES\n\t%s\n\t\t%s", c.confFile, "The system-wide configuration file")
		if c.confFlag != "" {
			uprint(` (overridden by --%s)`, c.confFlag)
		}
		uprint("\n\n")
	}
	uprint("ENVIRONMENT\n")
	for _, f := range fields {
		if f.flagName == c.confFlag {
			continue
		}
		_, help := unquoteHelp(&f)
		uprint("\t%s", f.envName)
		if f.boolField {
			uprint(" <true|false>")
		}
		if help != "" {
			uprint("\n\t\t%s", help)
		}
		uprint("\n")
	}
}

func uprint(s string, vals ...interface{}) {
	fmt.Fprintf(os.Stderr, s, vals...)
}

// unquoteUsage extracts a back-quoted name from the usage
// string for a flag and returns it and the un-quoted usage.
// Given "a `name` to show" it returns ("name", "a name to show").
// If there are no back quotes, the name is an educated guess of the
// type of the flag's value, or the empty string if the flag is boolean.
// (adapted from package flag)
func unquoteHelp(f *field) (name string, usage string) {
	// Look for a back-quoted name, but avoid the strings package.
	usage = f.options.help
	for i := 0; i < len(usage); i++ {
		if usage[i] == '\'' {
			for j := i + 1; j < len(usage); j++ {
				if usage[j] == '\'' {
					name = usage[i+1 : j]
					usage = usage[:i] + name + usage[j+1:]
				}
			}
			break // Only one single quote; use type name.
		}
	}

	if name == "" {
		// No explicit name, so use type if we can find one.
		name = "value"

		switch f.field.Kind() {
		case reflect.Bool:
			name = ""
		//TODO: duration
		//TODO: slice
		case reflect.Float32, reflect.Float64:
			name = "float"
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			typ := f.field.Type()
			if typ.PkgPath() == "time" && typ.Name() == "Duration" {
				name = "duration"
			} else {
				name = "int"
			}
		case reflect.String:
			name = "string"
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			name = "uint"
		}
	}
	if name != "" {
		name = `<` + name + `>`
	}
	return
}
