package source

import (
	"errors"
	"fmt"
	"strings"
)

// ErrHelpWanted provides an indication help was requested.
var ErrHelpWanted = errors.New("help wanted")

// Flag is a source for command line arguments.
type Flag struct {
	m map[string]string
}

// NewFlag parsing a string of command line arguments. NewFlag will return
// ErrHelpWanted, if the help flag is identifyed. This code is adapted
// from the Go standard library flag package.
func NewFlag(args []string) (*Flag, error) {
	m := make(map[string]string)

	if len(args) != 0 {
		for {
			if len(args) == 0 {
				break
			}

			// Look at the next arg.
			s := args[0]

			// If it's too short or doesn't begin with a `-`, assume we're at
			// the end of the flags.
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
				return nil, fmt.Errorf("bad flag syntax: %s", s)
			}

			// It's a flag. Does it have an argument?
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
				return nil, ErrHelpWanted
			}

			// If we don't have a value yet, it's possible the flag was not in the
			// -flag=value format which means it might still have a value which would be
			// the next argument, provided the next argument isn't a flag.
			if !hasValue {
				if len(args) > 0 && args[0][0] != '-' {

					// Doesn't look like a flag. Must be a value.
					value, args = args[0], args[1:]
				} else {

					// We assume this is a boolean flag.
					value = "true"
				}
			}

			// Store the flag/value pair.
			m[name] = value
		}
	}

	return &Flag{m: m}, nil
}

// Get implements the confg.Source interface. Returns the stringfied value
// stored at the specified key from the flag source.
func (f *Flag) Get(key []string) (string, bool) {
	k := strings.ToLower(strings.Join(key, `-`))
	val, found := f.m[k]
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
