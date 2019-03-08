package conf

import (
	"errors"
	"fmt"
	"os"
)

type flagSource struct {
	found map[string]string
}

var errHelpWanted = errors.New("help wanted")

// TODO?: make missing flags optionally throw error
func newFlagSource(fields []field, exempt []string) (*flagSource, []string, error) {
	found := make(map[string]string, len(fields))
	expected := make(map[string]*field, len(fields))
	shorts := make(map[string]string, len(fields))
	exemptFlags := make(map[string]struct{}, len(exempt))

	// some flags are special, like for specifying a config file flag, which
	// we definitely want to inspect, but don't represent field data
	for _, exemptFlag := range exempt {
		if exemptFlag != "" {
			exemptFlags[exemptFlag] = struct{}{}
		}
	}

	for i, field := range fields {
		expected[field.flagName] = &fields[i]
		if field.options.short != 0 {
			shorts[string(field.options.short)] = field.flagName
		}
	}

	args := make([]string, len(os.Args)-1)
	copy(args, os.Args[1:])

	if len(args) != 0 {
		//adapted from 'flag' package
		for {
			if len(args) == 0 {
				break
			}
			// look at the next arg
			s := args[0]
			// if it's too short or doesn't begin with a `-`, assume we're at the end of the flags
			if len(s) < 2 || s[0] != '-' {
				break
			}
			numMinuses := 1
			if s[1] == '-' {
				numMinuses++
				if len(s) == 2 { // "--" terminates the flags
					args = args[1:]
					break
				}
			}
			name := s[numMinuses:]
			if len(name) == 0 || name[0] == '-' || name[0] == '=' {
				return nil, nil, fmt.Errorf("bad flag syntax: %s", s)
			}

			// it's a flag. does it have an argument?
			args = args[1:]
			hasValue := false
			value := ""
			for i := 1; i < len(name); i++ { // equals cannot be first
				if name[i] == '=' {
					value = name[i+1:]
					hasValue = true
					name = name[0:i]
					break
				}
			}
			if name == "help" || name == "h" || name == "?" {
				return nil, nil, errHelpWanted
			}

			if long, ok := shorts[name]; ok {
				name = long
			}

			if expected[name] == nil {
				if _, ok := exemptFlags[name]; !ok {
					return nil, nil, fmt.Errorf("flag provided but not defined: -%s", name)
				}
			}

			// if we don't have a value yet, it's possible the flag was not in the
			// -flag=value format which means it might still have a value which would be
			// the next argument, provided the next argument isn't a flag
			if !hasValue {
				if len(args) > 0 && args[0][0] != '-' {
					// doesn't look like a flag. Must be a value
					value, args = args[0], args[1:]
				} else {
					// we wanted a value but found the end or another flag. The only time this is okay
					// is if this is a boolean flag, in which case `-flag` is okay, because it is assumed
					// to be the same as `-flag true`
					if expected[name].boolField {
						value = "true"
					} else {
						return nil, nil, fmt.Errorf("flag needs an argument: -%s", name)
					}
				}
			}
			found[name] = value
		}
	}

	return &flagSource{
		found: found,
	}, args, nil
}

func (f *flagSource) Get(key []string) (string, bool) {
	flagStr := getFlagName(key)
	val, found := f.found[flagStr]
	return val, found
}

/*
Portions Copyright (c) 2009 The Go Authors. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

   * Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
   * Redistributions in binary form must reproduce the above
copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.
   * Neither the name of Google Inc. nor the names of its
contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/
