package conf

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"text/tabwriter"
)

func printUsage(fields []field) {

	// Sort the fields by their long name.
	sort.SliceStable(fields, func(i, j int) bool {
		return fields[i].flagName < fields[j].flagName
	})

	fields = append(fields, field{
		flagName:  "help",
		boolField: true,
		options: fieldOptions{
			short: 'h',
			help:  "display this help message",
		}})

	fmt.Fprintf(os.Stderr, "Usage: %s [options] [arguments]\n\n", os.Args[0])

	fmt.Fprintln(os.Stderr, "OPTIONS")
	w := new(tabwriter.Writer)
	w.Init(os.Stderr, 0, 4, 2, ' ', tabwriter.TabIndent)

	for _, f := range fields {
		typeName, help := getTypeAndHelp(&f)
		fmt.Fprintf(w, "  --%s", f.flagName)
		if f.options.short != 0 {
			fmt.Fprintf(w, "/-%s", string(f.options.short))
		}
		if f.envName != "" {
			fmt.Fprintf(w, "/$%s", f.envName)
		}
		fmt.Fprintf(w, " %s\t%s\t\n", typeName, getOptString(f))
		if help != "" {
			fmt.Fprintf(w, "      %s\t\t\n", help)
		}
	}

	w.Flush()
	fmt.Fprintf(os.Stderr, "\n")
}

// getTypeAndHelp extracts the type and help message for a single field for
// printing in the usage message. If the help message contains text in
// single quotes ('), this is assumed to be a more specific "type", and will
// be returned as such. If there are no back quotes, it attempts to make a
// guess as to the type of the field. Boolean flags are not printed with a
// type, manually-specified or not, since their presence is equated with a
// 'true' value and their absence with a 'false' value. If a type cannot be
// determined, it will simply give the name "value". Slices will be annotated
// as "<Type>,[Type...]", where "Type" is whatever type name was chosen.
// (adapted from package flag).
func getTypeAndHelp(f *field) (name string, usage string) {

	// Look for a single-quoted name.
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

	var isSlice bool
	if f.field.IsValid() {
		t := f.field.Type()

		// If it's a pointer, we want to deref.
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		// If it's a slice, we want the type of the slice elements.
		if t.Kind() == reflect.Slice {
			t = t.Elem()
			isSlice = true
		}

		// If no explicit name was provided, attempt to get the type
		if name == "" {
			switch t.Kind() {
			case reflect.Bool:
				if !isSlice {
					return "", usage
				}
				name = ""
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
			default:
				name = "value"
			}
		}
	}

	switch {
	case isSlice:
		name = fmt.Sprintf("<%s>,[%s...]", name, name)
	case name != "":
		name = fmt.Sprintf("<%s>", name)
	default:
	}
	return
}

func getOptString(f field) string {
	opts := make([]string, 0, 3)
	if f.options.required {
		opts = append(opts, "required")
	}
	if f.options.noprint {
		opts = append(opts, "noprint")
	}
	if f.options.defaultStr != "" {
		opts = append(opts, fmt.Sprintf("default: %s", f.options.defaultStr))
	}
	if len(opts) > 0 {
		return fmt.Sprintf("(%s)", strings.Join(opts, `,`))
	}
	return ""
}
