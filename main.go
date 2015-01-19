
package main

import (
	"fmt"
	"github.com/Leimy/icy-metago/bot"
	"github.com/Leimy/icy-metago/shout"
	"log"
	"os"
	"flag"
)

var irc       = flag.String("server", "irc.radioxenu.com:6667", "<server>:<port>")
var room      = flag.String("channel", "#radioxenu", "<#channelname>")
var streamurl = flag.String("stream", "http://radioxenu.com:8000/relay", "<http://url-to-get-metadata>")


func usage(pname string) {
	fmt.Printf("%s irc.radioxenu.com:6667 http://radioxenu.com:8000/relay metabot #radioxenu\n", pname)
	os.Exit(0)
}

// maybe use flags
func main() {
	flag.Parse()

	botChan := make(chan string)
	for {
		theBot, err := bot.NewBot(
			*room,
			"metabot",
			*irc,
			botChan)

		if err != nil {
			log.Panic(err)
		}
		
		shout.GetMeta(*streamurl, theBot, botChan)

	}
}
