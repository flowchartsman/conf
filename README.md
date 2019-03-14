# conf

Simple, self-documenting, struct-driven configuration with flag generation and zero dependencies.

## Overview
`conf` provides a simple method to drive structured configuration from types and fields, with automatic flag and usage generation.

## Usage
```go
package main

import (
	"log"
	"time"

	"github.com/flowchartsman/conf"
)

type myConfig struct {
	Sub        subConfig
	TimeToWait time.Duration `conf:"help:how long to wait,short:c,required"`
	Password   string        `conf:"help:the database password to use,noprint"`
	DNSServer  *string       `conf:"help:the address of the dns server to use,default:127.0.0.1"`
	Debug      bool          `conf:"help:enable debug mode"`
	DBServers  []string      `conf:"help:a list of mirror 'host's to contact"`
}

type subConfig struct {
	Value int `conf:"help: I am a subvalue"`
}

func main() {
	log.SetFlags(0)
	var c myConfig
	err := conf.Parse(&c,
		conf.WithConfigFile("/etc/test.conf"),
		conf.WithConfigFileFlag("conf"))
	if err != nil {
		log.Fatal(err)
	}
	log.Println(conf.String(&c))
}
```

```
$ ./conftest -h
Usage: ./conftest [options] [arguments]

OPTIONS
  --db-servers <host>,[host...]                  DB_SERVERS
      a list of mirror hosts to contact
  --debug enable debug mode                      DEBUG
  --dns-server <string>                          DNS_SERVER
      the address of the dns server to use
      (default: 127.0.0.1)
  --password <string>                            PASSWORD
      the database password to use
      (noprint)
  --sub-value <int>                              SUB_VALUE
      I am a subvalue
  --time-to-wait, -c <duration>                  TIME_TO_WAIT
      how long to wait
      (required)
  --conf filename
      the filename to load configuration from
      (default: /etc/test.conf)
  --help, -h display this help message

FILES
  /etc/test.conf
    The system-wide configuration file (overridden by --conf)

$ ./conftest
required field TimeToWait is missing value
$ ./conftest --time-to-wait 5s --sub-value 1 --password I4mInvisbl3! --db-servers 127.0.0.1,127.0.0.2 --dns-server 1.1.1.1
SUB_VALUE=1 TIME_TO_WAIT=5s DNS_SERVER=1.1.1.1 DEBUG=false DB_SERVERS=[127.0.0.1 127.0.0.2] <nil>
```

## note
This library is still in **alpha**. It needs docs, full coverage testing, and poking to find edgecases.

## shoulders
This library takes inspiration (and some code) from some great work by some great engineers. These are credited in the license, but more detail soon.
- [kelseyhightower/envconfig](https://github.com/kelseyhightower/envconfig)
- [peterbourgon/ff](https://github.com/peterbourgon/ff)
