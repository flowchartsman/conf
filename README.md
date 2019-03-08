# conf

conf provides an opinionated, struct-first way of reading configuration

```go
package main

import (
        "fmt"
        "github.com/flowchartsman/conf"
        "time"
)

type EmbeddedConf struct {
        Baz int `conf:"help:the number of bazzes to baz"`
}

type ConfigTest struct {
        EmbeddedConf
        TimeToWait time.Duration `conf:"short:t,help:how long to wait"`
        DNSServer  string        `conf:"default:127.0.0.1,help:the 'address' of the dns server to use"`
        Thing      bool          `conf:"help:to thing or not to thing"`
}

func main() {
        var c ConfigTest
        err := conf.Parse(&c,
                conf.WithConfigFileFlag("conf"),
                conf.WithConfigFile("/etc/foo.conf"),
        )
        if err != nil {
                fmt.Println(err)
        }
        fmt.Printf("%#v\n", c)
}
```

```
./conftest -?
Usage: ./conftest [options] [arguments]

OPTIONS
        --baz <int>
                the number of bazzes to baz
        --conf <filename>
                the filename to load configuration from
                (default: /etc/foo.conf)
        --dns-server <address>
                the address of the dns server to use
                (default: 127.0.0.1)
        --thing to thing or not to thing
        --time-to-wait, -t <duration>
                how long to wait

FILES
        /etc/foo.conf
                The system-wide configuration file (overridden by --conf)

ENVIRONMENT
        BAZ
                the number of bazzes to baz
        DNS_SERVER
                the address of the dns server to use
        THING <true|false>
                to thing or not to thing
        TIME_TO_WAIT
                how long to wait
```

```
$ BAZ=1 ./conftest -t 5s -thing
main.ConfigTest{EmbeddedConf:main.EmbeddedConf{Baz:1}, TimeToWait:5000000000, DNSServer:"127.0.0.1", Thing:true}
```
