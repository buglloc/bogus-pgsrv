package main

import (
	"flag"
	"time"

	"github.com/buglloc/simplelog"

	"github.com/buglloc/bogus-pgsrv/pkg/pgsrv"
)

func main() {
	var addr string
	var breakTime int
	flag.StringVar(&addr, "addr", ":5432", "addr to listen")
	flag.IntVar(&breakTime, "break-time", 1000, "break time (ms)")
	flag.Parse()

	sleepTime := time.Duration(breakTime) * time.Millisecond
	srv := pgsrv.New()
	for {
		err := srv.Listen(addr)
		if err != nil {
			log.Error("loop done", "err", err.Error())
			time.Sleep(sleepTime)
		}
	}
}
