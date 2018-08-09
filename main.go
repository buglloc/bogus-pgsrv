package main

import (
	"flag"
	"time"

	"github.com/buglloc/simplelog"

	"github.com/buglloc/bogus-pgsrv/pkg/pgsrv"
)

var (
	addr string
)

func main() {
	flag.StringVar(&addr, "addr", ":5432", "addr to listen")
	flag.Parse()

	srv := pgsrv.New()

	for {
		err := srv.Listen(addr)
		if err != nil {
			log.Error(err.Error())
			time.Sleep(time.Second)
		}
	}
}
