package main

import (
	"fmt"
	"github.com/Leimy/icy-metago/shout"
	"os"
)

func usage(pname string) {
	fmt.Printf("%s http://radioxenu.com:8000/relay\n", pname)
	os.Exit(0)
}

func main() {
	if len(os.Args) < 2 {
		usage(os.Args[0])
	}

	for {
		shout.StreamMeta(os.Args[1])
	}
}
