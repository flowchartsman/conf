package conf

import (
	"io/ioutil"
	"os"
	"testing"
)

type simpleConf struct {
	TestInt    int
	TestString string
	TestBool   bool
}

func TestSimpleParseFlags(t *testing.T) {
	prepArgs(
		"--test-int", "1",
		"--test-string", "s",
		"--test-bool")
	prepEnv()
	var c simpleConf
	err := Parse(&c)
	assert(t, err == nil)
	assert(t, c.TestInt == 1)
	assert(t, c.TestString == "s")
	assert(t, c.TestBool)
}

func TestSimpleParseEnv(t *testing.T) {
	prepArgs()
	prepEnv(
		"TEST_INT", "1",
		"TEST_STRING", "s",
		"TEST_BOOL", "TRUE",
	)
	var c simpleConf
	err := Parse(&c)
	assert(t, err == nil)
	assert(t, c.TestInt == 1)
	assert(t, c.TestString == "s")
	assert(t, c.TestBool)
}

func TestSimpleFile(t *testing.T) {
	prepArgs()
	prepEnv()
	testFile, err := ioutil.TempFile("", "conf-test")
	if err != nil {
		panic("error creating temp file for test: " + err.Error())
	}
	defer os.Remove(testFile.Name())
	testFile.Write([]byte(`TEST_INT 1
TEST_STRING s
TEST_BOOL TRUE
`))
	err = testFile.Close()
	if err != nil {
		panic("error closing temp file for test: " + err.Error())
	}
	var c simpleConf
	err = Parse(&c,
		WithConfigFile(testFile.Name()),
	)
	assert(t, err == nil)
	assert(t, c.TestInt == 1)
	assert(t, c.TestString == "s")
	assert(t, c.TestBool)
}

func TestSimpleSourcePriority(t *testing.T) {
	type simpleConfPriority struct {
		TestInt      int
		TestIntTwo   int
		TestIntThree int
	}
	prepEnv(
		"TEST_INT", "1",
		"TEST_INT_TWO", "1",
		"TEST_INT_THREE", "1",
	)
	testFile, err := ioutil.TempFile("", "conf-test")
	if err != nil {
		panic("error creating temp file for test: " + err.Error())
	}
	defer os.Remove(testFile.Name())
	testFile.Write([]byte(`TEST_INT_TWO 2
TEST_INT_THREE 2
	`))
	err = testFile.Close()
	if err != nil {
		panic("error closing temp file for test: " + err.Error())
	}
	prepArgs(
		"--test-int-three", "3",
	)
	var c simpleConfPriority
	err = Parse(&c,
		WithConfigFile(testFile.Name()),
	)
	assert(t, err == nil)
	assert(t, c.TestInt == 1)
	assert(t, c.TestIntTwo == 2)
	assert(t, c.TestIntThree == 3)
}

func TestParseNonRefIsError(t *testing.T) {
	prepArgs()
	prepEnv()
	var c simpleConf
	err := Parse(c)
	assert(t, err == ErrInvalidStruct)
}

func TestParseNonStructIsError(t *testing.T) {
	prepArgs()
	prepEnv()
	var s string
	err := Parse(&s)
	assert(t, err == ErrInvalidStruct)
}

func TestSkipedFieldIsSkipped(t *testing.T) {
	type skipTest struct {
		TestString string `conf:"-"`
		TestInt    int
	}
	var c skipTest
	prepArgs()
	prepEnv(
		"TEST_STRING", "no",
		"TEST_INT", "1",
	)
	err := Parse(&c)

	assert(t, err == nil)
	assert(t, c.TestString == "")
	assert(t, c.TestInt == 1)
}

func TestTagMissingValueIsError(t *testing.T) {
	type bad struct {
		TestBad string `conf:"default:"`
	}
	var c bad
	prepArgs()
	prepEnv()
	err := Parse(&c)

	assert(t, err.Error() == `conf: error parsing tags for field TestBad: tag "default" missing a value`)
}

func TestBadShortTagIsError(t *testing.T) {
	type badShort struct {
		TestBad string `conf:"short:ab"`
	}
	var c badShort
	prepArgs()
	prepEnv()
	err := Parse(&c)

	assert(t, err.Error() == `conf: error parsing tags for field TestBad: short value must be a single rune, got "ab"`)
}

func TestCannotSetRequiredAndDefaultTags(t *testing.T) {
	type badShort struct {
		TestBad string `conf:"required,default:n"`
	}
	var c badShort
	prepArgs()
	prepEnv()
	err := Parse(&c)

	assert(t, err.Error() == "conf: error parsing tags for field TestBad: cannot set both `required` and `default`")
}

func TestRequiredMustBePresent(t *testing.T) {
	type required struct {
		NeededValue string `conf:"required"`
	}
	var c required
	prepArgs()
	prepEnv()
	err := Parse(&c)
	assert(t, err.Error() == "required field NeededValue is missing value")
}

func TestHierarchicalFieldNames(t *testing.T) {
	type conf1 struct {
		FieldOne string
	}
	type conf2 struct {
		One      conf1
		FieldTwo string
	}
	var c conf2
	prepArgs("--one-field-one=1")
	prepEnv("FIELD_TWO", "2")
	err := Parse(&c)
	assert(t, err == nil)
	assert(t, c.One.FieldOne == "1")
	assert(t, c.FieldTwo == "2")
}

func TestEmbeddedFieldNames(t *testing.T) {
	type Conf1 struct {
		FieldOne string
	}
	type conf2 struct {
		Conf1
		FieldTwo string
	}
	var c conf2

	prepEnv("FIELD_ONE", "1")
	prepArgs("--field-two=2")
	err := Parse(&c)
	assert(t, err == nil)
	assert(t, c.FieldOne == "1")
	assert(t, c.FieldTwo == "2")
}

func prepEnv(keyvals ...string) {
	if len(keyvals)%2 != 0 {
		panic("prepENV must have even number of keyvals")
	}
	os.Clearenv()
	for i := 0; i < len(keyvals); i += 2 {
		os.Setenv(keyvals[i], keyvals[i+1])
	}
}

func prepArgs(args ...string) {
	os.Args = append([]string{"testing"}, args...)
}

func assert(t *testing.T, testresult bool) {
	t.Helper()
	if !testresult {
		t.Fatal()
	}
}
