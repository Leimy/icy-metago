package main

import (
	"net/http"
	"log"
	"io"
	"fmt"
	"bufio"
	"os"
	"net"
	"strings"
)

type icyparseerror struct {
	s string
}

func (e *icyparseerror) Error() string {
	return e.s
}

func parseIcy(rdr *bufio.Reader, c byte) (string, error) {
	numbytes := int(c) * 16
	bytes := make([] byte, numbytes)
	n, err :=  io.ReadFull(rdr, bytes)
	if err != nil {
		log.Panic(err)
	}
	if n != numbytes {
		return "", &icyparseerror{"didn't get enough data"}  // may be invalid
	}
	return string(bytes), nil
}

func extractMetadata(rdr io.Reader, skip int) (<- chan string) {
	ch := make(chan string)
	go func () {
		bufrdr := bufio.NewReaderSize(rdr, skip)
		for {
			skipbytes := make([] byte, skip)

			_, err := io.ReadFull(bufrdr, skipbytes)
			if err != nil {
				log.Panic(err)
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

func shoutStreamStuff(url string) (<- chan string) {
	log.Printf("Shoutcast stream metadata yanker v0.1\n");
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
	_, err = fmt.Sscan(resp.Header.Get("Icy-Metaint"), &amount)
	if err != nil {
		log.Panic(err)
	}

	return extractMetadata(resp.Body, amount)
}	


func ircStreamStuff(conn net.Conn, botname, channel string, mdc <- chan string) {
	bior := bufio.NewReader(conn)
	biow := bufio.NewWriter(conn)

	fmt.Fprintf(biow, "NICK  %s\r\n", botname)
	fmt.Fprintf(biow, "USER %s 0 * :tutorial bot\r\n", botname)
	fmt.Fprintf(biow, "JOIN %s\r\n", channel)

	cc := make(chan string)
	go func () {
		for { 
			bytes, morep, err := bior.ReadLine()
			if err != nil {
				log.Panic(err)
			}
			for morep {
				bytes2, morep, err := bior.ReadLine()
				if err != nil {
					log.Panic(err)
				}
				bytes = append(bytes, bytes2...)
				if !morep {
					break
				}
			}
			cc <- string(bytes)
		}
	} ()
			
	for {
		if err := biow.Flush(); err != nil {
			log.Panic(err)
		}
		select {
		case input:= <- mdc:
			fmt.Fprintf(biow, "PRIVMSG %s : Now Playing: %s\r\n", channel, input)
		case line := <- cc:
			log.Print(line)
				
			// Handle server ping, so we don't get booted.
			if strings.HasPrefix(line, "PING :") {
				resp := strings.Replace(line, "PING", "PONG", 1)
				fmt.Fprintf(biow, "%s\r\n", resp)
			}
		}
	}
}

func bot(serveraddy string, botname, channel string, metadata <- chan string) {
	log.Print("Connecting to ", serveraddy)
	conn, err := net.Dial("tcp4", serveraddy)
	if err != nil {
		log.Panic(err)
	}
	log.Print("Done")

	ircStreamStuff(conn, botname, channel, metadata)
}

func usage () {
	fmt.Printf("./shoutmeta irc.radioxenu.com:8000/relay irc.radioxenu.com:6667 metabot #radioxenu\n")
	os.Exit(0)
}

func main () {
	if len(os.Args) < 5 {
		log.Print(len(os.Args))
		usage()
	}

	bot(os.Args[2], os.Args[3], os.Args[4], shoutStreamStuff(os.Args[1]))
	log.Panic("Stream broke... dangit!")
}
