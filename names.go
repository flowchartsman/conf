package conf

import (
	"strings"
	"unicode"
)

func getEnvName(key []string) string {
	return strings.ToUpper(strings.Join(key, `_`))
}

func getFlagName(key []string) string {
	return strings.ToLower(strings.Join(key, `-`))
}

// split string based on camel case
func camelSplit(src string) []string {
	if src == "" {
		return []string{}
	}
	if len(src) < 2 {
		return []string{src}
	}

	runes := []rune(src)

	lastClass := charClass(runes[0])
	lastIdx := 0
	out := []string{}

	// split into fields based on class of unicode character
	for i, r := range runes {
		class := charClass(r)
		// if the class has transitioned
		if class != lastClass {
			// if going from uppercase to lowercase, we want to retain the last
			// uppercase letter for names like FOOBar, which should split to
			// FOO Bar
			if lastClass == classUpper && class != classNumber {
				if i-lastIdx > 1 {
					out = append(out, string(runes[lastIdx:i-1]))
					lastIdx = i - 1
				}
			} else {
				out = append(out, string(runes[lastIdx:i]))
				lastIdx = i
			}
		}

		if i == len(runes)-1 {
			out = append(out, string(runes[lastIdx:]))
		}
		lastClass = class

	}

	return out
}

const (
	classLower int = iota
	classUpper
	classNumber
	classOther
)

func charClass(r rune) int {
	switch {
	case unicode.IsLower(r):
		return classLower
	case unicode.IsUpper(r):
		return classUpper
	case unicode.IsDigit(r):
		return classNumber
	}
	return classOther
}
