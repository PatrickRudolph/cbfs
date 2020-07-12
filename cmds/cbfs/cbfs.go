package main

import (
	"fmt"
	"log"

	"github.com/PatrickRudolph/cbfs/pkg/cbfs"
	flag "github.com/spf13/pflag"
)

var debug = flag.BoolP("debug", "d", false, "enable debug prints")

func main() {
	flag.Parse()

	if *debug {
		cbfs.Debug = log.Printf
	}

	a := flag.Args()
	if len(a) != 2 {
		log.Fatal("arg count")
	}

	i, err := cbfs.Open(a[0])
	if err != nil {
		log.Fatal("Failed to open file %s. Error:", a[0], err)
	}

	switch a[1] {
	case "list":
		fmt.Printf("%s", i.String())
	default:
		log.Fatal("?")
	}

}
