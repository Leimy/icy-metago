package shout

import (
	"bufio"
	"fmt"
	"github.com/Leimy/icy-metago/bot"
	"io"
	"log"
	"net/http"
	"strings"
	"github.com/Leimy/icy-metago/twitter"
)

// Auto settings stuff
const (
	noOperation = iota
	toggleTweet
	toggleLast
	get
)

type autoSettings struct {
	Op int
	Tweet bool
	Last bool
}

func startAutoSettings() (chan autoSettings) {
	settings := make(chan autoSettings)
	go func() {
		curSettings := &autoSettings{}
		for {
			s := <- settings
			switch (s.Op) {
			case noOperation:
				return
			case toggleTweet:
				curSettings.Tweet = !curSettings.Tweet
			case toggleLast:
				curSettings.Last = !curSettings.Last
			case get:
				settings <- *curSettings
			}
		}
	}()

	return settings
}

func toggleAutoTweet(settings chan autoSettings) {
	settings <- autoSettings{toggleTweet, false, false}
}

func toggleAutoLast(settings chan autoSettings) {
	settings <- autoSettings{toggleLast, false, false}
}

func getAutoSettings(settings chan autoSettings)  autoSettings {
	settings <- autoSettings{get, false, false}
	return <- settings
}
// end end auto settings

type icyparseerror struct {
	s string
}

func (ipe *icyparseerror) Error() string {
	return ipe.s
}

func parseIcy(rdr *bufio.Reader, c byte) (string, error) {
	numbytes := int(c) * 16
	bytes := make([]byte, numbytes)
	n, err := io.ReadFull(rdr, bytes)
	if err != nil {
		log.Panic(err)
	}
	if n != numbytes {
		return "", &icyparseerror{"didn't get enough data"} // may be invalid
	}
	return strings.Split(strings.Split(string(bytes), "=")[1], ";")[0], nil
}

func extractMetadata(rdr io.Reader, skip int) <-chan string {
	ch := make(chan string)
	go func() {
		bufrdr := bufio.NewReaderSize(rdr, skip)
		for {
			skipbytes := make([]byte, skip)

			_, err := io.ReadFull(bufrdr, skipbytes)
			if err != nil {
				log.Printf("Failed: %v\n", err)
				close(ch)
				break;
			}
			c, err := bufrdr.ReadByte()
			if err != nil {
				log.Panic(err)
			}
			if c > 0 {
				meta, err := parseIcy(bufrdr, c)
				if err != nil {
					log.Panic(err)
				}
				ch <- meta
			}
		}
	}()
	return ch
}

func StreamMeta(url string) {
	log.Printf("Shoutcast stream metadata yanker v0.1\n")
	client := &http.Client{}

	log.Printf("Getting from : %s\n", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("%v\n", err)
		return
	}

	req.Header.Add("Icy-MetaData", "1")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("%v\n", err)
		return
	}

	amount := 0
	if _, err = fmt.Sscan(resp.Header.Get("Icy-Metaint"), &amount); err != nil {
		log.Printf("%v\n", err)
		return
	}

	metaChan := extractMetadata(resp.Body, amount)

	for meta := range metaChan {
		fmt.Printf("%s\n", meta)
	}		
}

func GetMeta(url string, bot *bot.Bot, requestChan chan string) {
	log.Printf("Shoutcast stream metadata yanker v0.5\n")
	client := &http.Client{}

	log.Printf("Getting from : %s\n", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("%v\n", err)
		return
	}

	req.Header.Add("Icy-MetaData", "1")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("%v\n", err)
		return
	}

	amount := 0
	if _, err = fmt.Sscan(resp.Header.Get("Icy-Metaint"), &amount); err != nil {
		log.Printf("%v\n", err)
		return
	}

	metaChan := extractMetadata(resp.Body, amount)

	var lastsong string
	settings := startAutoSettings()
	
	for {
		select {
		case lastsong = <-metaChan:
			if lastsong == "" {
				return
			}
			auto_settings := getAutoSettings(settings)
			if  auto_settings.Last {
				bot.StringReplyCommand(fmt.Sprintf("Now Playing: %s", lastsong))
			}
			
			if auto_settings.Tweet {
				twitter.Tweet(lastsong)
			}
					
		case request := <-requestChan:
			switch (request) {
			case "?autotweet?":
				toggleAutoTweet(settings)
				atonoff := getAutoSettings(settings)
				log.Printf("AUTOTWEET: %v\n", atonoff.Tweet)
				bot.StringReplyCommand(fmt.Sprintf("autotweeting is %v", atonoff.Tweet))
			case "?autolast?":
				toggleAutoLast(settings)
				alonoff := getAutoSettings(settings)
				log.Printf("AUTOLAST: %v\n", alonoff.Last)
				bot.StringReplyCommand(fmt.Sprintf("autolast is %v", alonoff.Last))
			case "?lastsong?":
				log.Printf("Got a request to print the metadata which is: %s\n", lastsong)
				bot.StringReplyCommand(lastsong)
			case "?tweet?":
				log.Printf("Got a request to tweet that meta (%s)\n", lastsong)
				twitter.Tweet(lastsong)
			case "":
				log.Printf("Bot died!, we're out too!")
				return
			}
		}
	}
}
