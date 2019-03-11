# conf

Simple, self-documenting, struct-driven configuration with flag generation and zero dependencies.

## Overview
`conf` provides a simple method to drive structured configuration from types and fields, with automatic flag and usage generation.

## Usage
```go
package main

import (
	"fmt"
	"github.com/flowchartsman/conf"
	"time"
)

type myConfig struct {
	TimeToWait time.Duration `conf:"short:t,help:how long to wait"`
	DNSServer  string        `conf:"default:127.0.0.1,help:the 'address' of the dns server to use"`
	Debug      bool          `conf:"help:enable debug mode"`
}

func main() {
	var c myConfig
	err := conf.Parse(&c)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%#v\n", c)
}
```

```
$ ./conftest -h
Usage: ./conftest [options] [arguments]

OPTIONS
        --debug enable debug mode
        --dns-server <address>
                the address of the dns server to use
                (default: 127.0.0.1)
        --help, -h      display this help message
        --time-to-wait, -t <duration>
                how long to wait

ENVIRONMENT
        DEBUG <true|false>
                enable debug mode
        DNS_SERVER
                the address of the dns server to use
        TIME_TO_WAIT
                how long to wait

$ ./conftest
main.myConfig{TimeToWait:0, DNSServer:"127.0.0.1", Debug:false}
$ export TIME_TO_WAIT=5s
$ ./conftest --dns-server=192.168.1.1 -debug
main.myConfig{TimeToWait:5000000000, DNSServer:"192.168.1.1", Debug:true}
```

## note
This library is still in **alpha**. It needs docs, full coverage testing (well, tests at all), and poking to find edgecases.

## shoulders
This library takes inspiration (and some code) from some great work by some great engineers. These are credited in the license, but more detail soon.
- [kelseyhightower/envconfig](https://github.com/kelseyhightower/envconfig)
- [peterbourgon/ff](https://github.com/peterbourgon/ff)
