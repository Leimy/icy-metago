package main

import (
	"fmt"
	"icy-metago/bot"
	"icy-metago/shout"
	"log"
	"os"
)

func usage(pname string) {
	fmt.Printf("%s irc.radioxenu.com:6667 http://radioxenu.com:8000/relay metabot #radioxenu\n", pname)
	os.Exit(0)
}

// maybe use flags
func main() {
	if len(os.Args) < 5 {
		log.Print(len(os.Args))
		usage(os.Args[0])
	}

	bot, err := bot.NewBot(
		os.Args[4],
		os.Args[3],
		os.Args[1])
	if err != nil {
		log.Panic(err)
	}

	shout.GetMeta(os.Args[2], bot)
}
