package conf

// Option represents a change to the default parsing.
type Option func(c *context)

// WithConfigFile tells parse to attempt to read from the specified file, if it
// is found.
func WithConfigFile(filename string) Option {
	return func(c *context) {
		c.confFile = filename
	}
}

// WithConfigFileFlag tells parse to look for a flag called `flagname` and, if
// it is found, to attempt to load configuration from this file. If the flag
// is specified, it will override the value provided to WithConfigFile, if that
// has been specified. If the file is not found, the program will exit with an
// error.
func WithConfigFileFlag(flagname string) Option {
	return func(c *context) {
		c.confFlag = flagname
	}
}

// WithSource adds additional configuration sources for configuration parsing.
func WithSource(source Source) Option {
	return func(c *context) {
		c.sources = append(c.sources, source)
	}
}
