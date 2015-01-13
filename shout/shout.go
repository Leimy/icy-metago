package shout

import (
	"bufio"
	"fmt"
	"github.com/Leimy/icy-metago/bot"
	"io"
	"log"
	"net/http"
	"strings"
)

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
		log.Panic(err)
	}

	req.Header.Add("Icy-MetaData", "1")
	resp, err := client.Do(req)
	if err != nil {
		log.Panic(err)
	}

	amount := 0
	if _, err = fmt.Sscan(resp.Header.Get("Icy-Metaint"), &amount); err != nil {
		log.Panic(err)
	}

	metaChan := extractMetadata(resp.Body, amount)

	for meta := range metaChan {
		fmt.Printf("%s\n", meta)
	}		
}

func GetMeta(url string, bot *bot.Bot) {
	log.Printf("Shoutcast stream metadata yanker v0.1\n")
	client := &http.Client{}

	log.Printf("Getting from : %s\n", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Panic(err)
	}

	req.Header.Add("Icy-MetaData", "1")
	resp, err := client.Do(req)
	if err != nil {
		log.Panic(err)
	}

	amount := 0
	if _, err = fmt.Sscan(resp.Header.Get("Icy-Metaint"), &amount); err != nil {
		log.Panic(err)
	}

	metaChan := extractMetadata(resp.Body, amount)

	var lastsong string
	for {
		select {
		case lastsong = <-metaChan:
			if lastsong == "" {
				return;
			}
		case request := <-bot.SChan:
			if request == "?lastsong?" {
				log.Printf("Got a request to print the metadata which is: %s\n", lastsong)
				bot.StringReplyCommand(lastsong)
			}
		}
	}

}
