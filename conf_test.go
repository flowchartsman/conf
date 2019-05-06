package conf_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/flowchartsman/conf"
	"github.com/flowchartsman/conf/source"
	"github.com/google/go-cmp/cmp"
)

const (
	success = "\u2713"
	failed  = "\u2717"
)

const (
	ENV = iota + 1
	FLAG
	FILE
)

// NewSource returns an initialized source for a given type.
func NewSource(src int, v interface{}) (conf.Source, error) {
	switch src {
	case ENV:
		vars := v.(map[string]string)
		os.Clearenv()
		for k, v := range vars {
			os.Setenv(k, v)
		}
		return source.NewEnv("TEST")

	case FLAG:
		args := v.([]string)
		return source.NewFlag(args)

	case FILE:
		d := v.(struct {
			file *os.File
			vars map[string]string
		})
		var vars string
		for k, v := range d.vars {
			vars += fmt.Sprintf("%s %s\n", k, v)
		}
		if _, err := d.file.WriteString(vars); err != nil {
			return nil, err
		}
		if err := d.file.Close(); err != nil {
			return nil, err
		}
		return source.NewFile(d.file.Name())
	}

	return nil, errors.New("invalid source provided")
}

func TestBasicParse(t *testing.T) {
	type config struct {
		TestInt    int
		TestString string
		TestBool   bool
	}

	tests := []struct {
		name string
		src  int
		args interface{}
	}{
		{"basic-flag", FLAG, []string{"--test-int", "1", "--test-string", "s", "--test-bool"}},
		{"basic-env", ENV, map[string]string{"TEST_INT": "1", "TEST_STRING": "s", "TEST_BOOL": "TRUE"}},
		{"basic-file", FILE, map[string]string{"TEST_INT": "1", "TEST_STRING": "s", "TEST_BOOL": "TRUE"}},
	}

	t.Log("Given the need to parse configuration.")
	{
		for i, tt := range tests {
			t.Logf("\tTest: %d\tWhen checking this %d with arguments %s", i, tt.src, tt.args)
			{
				f := func(t *testing.T) {
					var source conf.Source

					switch tt.src {
					case ENV, FLAG:
						var err error
						source, err = NewSource(tt.src, tt.args)
						if err != nil {
							t.Fatalf("\t%s\tShould be able to call NewFlag : %s.", failed, err)
						}
						t.Logf("\t%s\tShould be able to call NewFlag.", success)

					case FILE:
						tf, err := ioutil.TempFile("", "conf-test")
						if err != nil {
							t.Fatalf("\t%s\tShould be able to create a temp file : %s.", failed, err)
						}
						t.Logf("\t%s\tShould be able to create a temp file.", success)
						defer os.Remove(tf.Name())

						d := struct {
							file *os.File
							vars map[string]string
						}{
							file: tf,
							vars: tt.args.(map[string]string),
						}

						source, err = NewSource(tt.src, d)
						if err != nil {
							t.Fatalf("\t%s\tShould be able to call NewFlag : %s.", failed, err)
						}
						t.Logf("\t%s\tShould be able to call NewFlag.", success)
					}

					var cfg config
					if err := conf.Parse(&cfg, source); err != nil {
						t.Fatalf("\t%s\tShould be able to Parse arguments : %s.", failed, err)
					}
					t.Logf("\t%s\tShould be able to Parse arguments.", success)

					want := config{
						TestInt:    1,
						TestString: "s",
						TestBool:   true,
					}
					if diff := cmp.Diff(want, cfg); diff != "" {
						t.Fatalf("\t%s\tShould have properly initialized struct value\n%s", failed, diff)
					}
					t.Logf("\t%s\tShould have properly initialized struct value.", success)
				}

				t.Run(tt.name, f)
			}
		}
	}
}

func TestMultiSource(t *testing.T) {
}

func TestParseNonRefIsError(t *testing.T) {
}

func TestParseNonStructIsError(t *testing.T) {
}

func TestSkipedFieldIsSkipped(t *testing.T) {
}

func TestTagMissingValueIsError(t *testing.T) {
}

func TestBadShortTagIsError(t *testing.T) {
}

func TestCannotSetRequiredAndDefaultTags(t *testing.T) {
}

func TestHierarchicalFieldNames(t *testing.T) {
}

func TestEmbeddedFieldNames(t *testing.T) {
}
