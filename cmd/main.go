// Package main is the entry point of the service.
// It exposes an optional command line argument, through which the value of max number of cores can be set for the server to use.
package main

import (
	"flag"
	"fmt"
	"github.com/coderafting/sentiment-analysis/internal/service"
	"runtime"
)

const defaultMaxProcs = 4

func main() {
	maxProcs := flag.Int("p", defaultMaxProcs, "Number of parallel processes in the machine.")
	flag.Parse()
	runtime.GOMAXPROCS(*maxProcs)
	a := service.GetApp()
	fmt.Println("Starting server on port", a.Cf.Port)
	a.StartServer()
}
