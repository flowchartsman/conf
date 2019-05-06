package conf_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
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
	DEFAULT = iota
	ENV
	FLAG
	FILE
)

var srcNames = []string{"DEFAULT", "ENV", "FLAG", "FILE"}

// NewSource returns an initialized source for a given type.
func NewSource(src int, v interface{}) (conf.Sourcer, error) {
	switch src {
	case DEFAULT:
		return nil, nil

	case ENV:
		args := v.(map[string]string)
		os.Clearenv()
		for k, v := range args {
			os.Setenv(k, v)
		}
		return source.NewEnv("TEST")

	case FLAG:
		args := v.([]string)
		return source.NewFlag(args)

	case FILE:
		args := v.(map[string]string)
		tf, err := ioutil.TempFile("", "conf-test")
		if err != nil {
			return nil, err
		}
		defer os.Remove(tf.Name())
		var vars string
		for k, v := range args {
			vars += fmt.Sprintf("%s %s\n", k, v)
		}
		if _, err := tf.WriteString(vars); err != nil {
			return nil, err
		}
		if err := tf.Close(); err != nil {
			return nil, err
		}
		return source.NewFile(tf.Name())
	}

	return nil, errors.New("invalid source provided")
}

func TestParse(t *testing.T) {
	type ip struct {
		Name string `conf:"default:localhost"`
		IP   string `conf:"default:127.0.0.0"`
	}
	type Embed struct {
		Name string `conf:"default:bill"`
	}
	type config struct {
		AnInt   int    `conf:"default:9"`
		AString string `conf:"default:B,short:s"`
		Bool    bool
		Skip    string `conf:"-"`
		IP      ip
		Embed
	}

	tests := []struct {
		name string
		src  int
		args interface{}
		want config
	}{
		{"default", DEFAULT, nil, config{9, "B", false, "", ip{"localhost", "127.0.0.0"}, Embed{"bill"}}},
		{"env", ENV, map[string]string{"TEST_AN_INT": "1", "TEST_S": "s", "TEST_BOOL": "TRUE", "TEST_SKIP": "SKIP", "TEST_IP_NAME": "local", "TEST_NAME": "andy"}, config{1, "s", true, "", ip{"local", "127.0.0.0"}, Embed{"andy"}}},
		{"flag", FLAG, []string{"--an-int", "1", "-s", "s", "--bool", "--skip", "skip", "--ip-name", "local", "--name", "andy"}, config{1, "s", true, "", ip{"local", "127.0.0.0"}, Embed{"andy"}}},
		{"file", FILE, map[string]string{"AN_INT": "1", "S": "s", "BOOL": "TRUE", "SKIP": "skip", "IP_NAME": "local", "NAME": "andy"}, config{1, "s", true, "", ip{"local", "127.0.0.0"}, Embed{"andy"}}},
	}

	t.Log("Given the need to parse basic configuration.")
	{
		for i, tt := range tests {
			t.Logf("\tTest: %d\tWhen checking %s with arguments %v", i, srcNames[tt.src], tt.args)
			{
				f := func(t *testing.T) {
					sourcer, err := NewSource(tt.src, tt.args)
					if err != nil {
						t.Fatalf("\t%s\tShould be able to create a new %s source : %s.", failed, srcNames[tt.src], err)
					}
					t.Logf("\t%s\tShould be able to create a new %s source.", success, srcNames[tt.src])

					var cfg config
					if err := conf.Parse(&cfg, sourcer); err != nil {
						t.Fatalf("\t%s\tShould be able to Parse arguments : %s.", failed, err)
					}
					t.Logf("\t%s\tShould be able to Parse arguments.", success)

					if diff := cmp.Diff(tt.want, cfg); diff != "" {
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
	type config struct {
		AnInt   int    `conf:"default:9"`
		AString string `conf:"default:B,short:s"`
		Bool    bool
	}

	tests := []struct {
		name    string
		sources []struct {
			src  int
			args interface{}
		}
		want config
	}{
		{
			name: "basic-env-flag",
			sources: []struct {
				src  int
				args interface{}
			}{
				{ENV, map[string]string{"TEST_AN_INT": "1", "TEST_S": "s", "TEST_BOOL": "TRUE"}},
				{FLAG, []string{"--an-int", "2", "-s", "s", "--bool", "false"}},
			},
			want: config{2, "s", false},
		},
	}

	t.Log("Given the need to parse multi-source configurations.")
	{
		for i, tt := range tests {
			t.Logf("\tTest: %d\tWhen checking %d sources", i, len(tt.sources))
			{
				f := func(t *testing.T) {
					var cfg config

					sources := make([]conf.Sourcer, len(tt.sources))
					for i, ttt := range tt.sources {
						sourcer, err := NewSource(ttt.src, ttt.args)
						if err != nil {
							t.Fatalf("\t%s\tShould be able to create a new %s source : %s.", failed, srcNames[ttt.src], err)
						}
						t.Logf("\t%s\tShould be able to create a new %s source.", success, srcNames[ttt.src])
						sources[i] = sourcer
					}

					if err := conf.Parse(&cfg, sources...); err != nil {
						t.Fatalf("\t%s\tShould be able to Parse arguments : %s.", failed, err)
					}
					t.Logf("\t%s\tShould be able to Parse arguments.", success)

					if diff := cmp.Diff(tt.want, cfg); diff != "" {
						t.Fatalf("\t%s\tShould have properly initialized struct value\n%s", failed, diff)
					}
					t.Logf("\t%s\tShould have properly initialized struct value.", success)
				}

				t.Run(tt.name, f)
			}
		}
	}
}

func TestFlagParse(t *testing.T) {
	type config struct {
		AnInt   int    `conf:"required,short:i"`
		AString string `conf:"default:B"`
		Bool    bool   `conf:"default:true"`
	}

	tests := []struct {
		name string
		src  int
		args interface{}
		want config
	}{
		{"basic-flag", FLAG, []string{"-i", "1", "--a-string", "s", "--bool"}, config{1, "s", true}},
	}

	t.Log("Given the need to parse basic configuration.")
	{
		for i, tt := range tests {
			t.Logf("\tTest: %d\tWhen checking %s with arguments %s", i, srcNames[tt.src], tt.args)
			{
				f := func(t *testing.T) {
					sourcer, err := NewSource(tt.src, tt.args)
					if err != nil {
						t.Fatalf("\t%s\tShould be able to create a new %s source : %s.", failed, srcNames[tt.src], err)
					}
					t.Logf("\t%s\tShould be able to create a new %s source.", success, srcNames[tt.src])

					var cfg config
					if err := conf.Parse(&cfg, sourcer); err != nil {
						t.Fatalf("\t%s\tShould be able to Parse arguments : %s.", failed, err)
					}
					t.Logf("\t%s\tShould be able to Parse arguments.", success)

					if diff := cmp.Diff(tt.want, cfg); diff != "" {
						t.Fatalf("\t%s\tShould have properly initialized struct value\n%s", failed, diff)
					}
					t.Logf("\t%s\tShould have properly initialized struct value.", success)
				}

				t.Run(tt.name, f)
			}
		}
	}
}

func TestErrors(t *testing.T) {
	t.Log("Given the need to validate errors that can occur with Parse.")
	{
		t.Logf("\tTest: %d\tWhen passing bad values to Parse.", 0)
		{
			f := func(t *testing.T) {
				var cfg struct {
					TestInt    int
					TestString string
					TestBool   bool
				}
				err := conf.Parse(cfg)
				if err == nil {
					t.Fatalf("\t%s\tShould NOT be able to accept a value by value.", failed)
				}
				t.Logf("\t%s\tShould NOT be able to accept a value by value.", success)
			}
			t.Run("not-by-ref", f)

			f = func(t *testing.T) {
				var cfg []string
				err := conf.Parse(cfg)
				if err == nil {
					t.Fatalf("\t%s\tShould NOT be able to pass anything but a struct value.", failed)
				}
				t.Logf("\t%s\tShould NOT be able to pass anything but a struct value.", success)
			}
			t.Run("not-struct-value", f)
		}

		t.Logf("\tTest: %d\tWhen bad tags to Parse.", 1)
		{
			f := func(t *testing.T) {
				var cfg struct {
					TestInt    int `conf:"default:"`
					TestString string
					TestBool   bool
				}
				err := conf.Parse(&cfg)
				if err == nil {
					t.Fatalf("\t%s\tShould NOT be able to accept tag missing value.", failed)
				}
				t.Logf("\t%s\tShould NOT be able to accept tag missing value.", success)
			}
			t.Run("tag-missing-value", f)

			f = func(t *testing.T) {
				var cfg struct {
					TestInt    int `conf:"short:ab"`
					TestString string
					TestBool   bool
				}
				err := conf.Parse(&cfg)
				if err == nil {
					t.Fatalf("\t%s\tShould NOT be able to accept invalid short tag.", failed)
				}
				t.Logf("\t%s\tShould NOT be able to accept invalid short tag.", success)
			}
			t.Run("tag-bad-short", f)
		}

		t.Logf("\tTest: %d\tWhen required values are missing.", 2)
		{
			f := func(t *testing.T) {
				var cfg struct {
					TestInt    int `conf:"required, default:1"`
					TestString string
					TestBool   bool
				}
				err := conf.Parse(&cfg)
				if err == nil {
					t.Fatalf("\t%s\tShould fail for missing required value.", failed)
				}
				t.Logf("\t%s\tShould fail for missing required value.", success)
			}
			t.Run("required-missing-value", f)
		}
	}
}

func TestUsage(t *testing.T) {
	t.Log("Given the need validate usage output.")
	{
		t.Logf("\tTest: %d\tWhen using a basic struct.", 0)
		{
			type config struct {
				AnInt   int    `conf:"default:9"`
				AString string `conf:"default:B,short:s"`
				Bool    bool
				Skip    []float64 `conf:"-"`
			}

			test := struct {
				name string
				src  int
				args interface{}
			}{
				name: "basic-env",
				src:  ENV,
				args: map[string]string{"TEST_ANINT": "1", "TEST_S": "s", "TEST_BOOL": "TRUE"},
			}

			sourcer, err := NewSource(test.src, test.args)
			if err != nil {
				fmt.Print(err)
				return
			}

			var cfg config
			if err := conf.Parse(&cfg, sourcer); err != nil {
				fmt.Print(err)
				return
			}

			got, err := conf.Usage(&cfg)
			if err != nil {
				fmt.Print(err)
				return
			}

			got = strings.TrimRight(got, " \n")
			want := `Usage: conf.test [options] [arguments]

OPTIONS
  --a-string/-s/$a-string <string>  (default: B)  
  --an-int/$an-int <int>            (default: 9)  
  --bool/$bool                                    
  --help/-h                                       
      display this help message`

			bGot := []byte(got)
			bWant := []byte(want)
			if diff := cmp.Diff(bGot, bWant); diff != "" {
				t.Log("got:\n", got)
				t.Log("\n", bGot)
				t.Log("wait:\n", want)
				t.Log("\n", bWant)
				t.Fatalf("\t%s\tShould match byte for byte the output.", failed)
			}
			t.Logf("\t%s\tShould match byte for byte the output.", success)
		}
	}
}

func ExampleString() {
	type config struct {
		AnInt   int    `conf:"default:9"`
		AString string `conf:"default:B,short:s"`
		Bool    bool
		Skip    []float64 `conf:"-"`
	}

	test := struct {
		name string
		src  int
		args interface{}
	}{
		name: "basic-env",
		src:  ENV,
		args: map[string]string{"TEST_AN_INT": "1", "TEST_S": "s", "TEST_BOOL": "TRUE"},
	}

	sourcer, err := NewSource(test.src, test.args)
	if err != nil {
		fmt.Print(err)
		return
	}

	var cfg config
	if err := conf.Parse(&cfg, sourcer); err != nil {
		fmt.Print(err)
		return
	}

	out, err := conf.String(&cfg)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Print(out)

	// Output:
	// an-int=1 a-string=s bool=true
}
