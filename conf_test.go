package conf_test

import (
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

func TestParseFlags(t *testing.T) {
	type config struct {
		TestInt    int
		TestString string
		TestBool   bool
	}

	tests := []struct {
		name string
		args []string
	}{
		{"basic", []string{"--test-int", "1", "--test-string", "s", "--test-bool"}},
	}

	t.Log("Given the need to parse command line arguments.")
	{
		for i, tt := range tests {
			t.Logf("\tTest: %d\tWhen checking these arguments %s", i, tt.args)
			{
				flag, err := source.NewFlag(tt.args)
				if err != nil {
					t.Fatalf("\t%s\tShould be able to call NewFlag : %s.", failed, err)
				}
				t.Logf("\t%s\tShould be able to call NewFlag.", success)

				var cfg config
				err = conf.Parse(&cfg, flag)
				if err != nil {
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
		}
	}
}

func TestParseEnv(t *testing.T) {
	type config struct {
		TestInt    int
		TestString string
		TestBool   bool
	}

	tests := []struct {
		name string
		vars map[string]string
	}{
		{"basic", map[string]string{"TEST_INT": "1", "TEST_STRING": "s", "TEST_BOOL": "TRUE"}},
	}

	t.Log("Given the need to parse environmental variables.")
	{
		for i, tt := range tests {
			t.Logf("\tTest: %d\tWhen checking these environmental variables %s", i, tt.vars)
			{
				os.Clearenv()
				for k, v := range tt.vars {
					os.Setenv(k, v)
				}

				env, err := source.NewEnv("TEST")
				if err != nil {
					t.Fatalf("\t%s\tShould be able to call NewEnv : %s.", failed, err)
				}
				t.Logf("\t%s\tShould be able to call NewEnv.", success)

				var cfg config
				err = conf.Parse(&cfg, env)
				if err != nil {
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
		}
	}
}

func TestParseFile(t *testing.T) {
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
