package main

import (
	"net/http"
	"log"
	"io"
	"fmt"
	"bufio"
	"os"
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
		log.Fatal(err)
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
				log.Fatal(err)
			}
			c, err := bufrdr.ReadByte()
			if err != nil {
				log.Fatal(err)
			}
			if c > 0 {
				meta, err := parseIcy(bufrdr, c)
				if err != nil {
					log.Fatal(err)
				}
				ch <- meta
			} 				
		}
	}()
	return ch
}

func main () {
	log.Printf("Shoutcast stream metadata yanker v0.1\n");
	client := &http.Client{}
	
	req, err := http.NewRequest("GET", os.Args[1], nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("Icy-MetaData", "1")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	amount := 0
	_, err = fmt.Sscan(resp.Header.Get("Icy-Metaint"), &amount)
	if err != nil {
		log.Fatal(err)
	}

	c := extractMetadata(resp.Body, amount)
	for meta := range(c) {
		fmt.Println(string(meta))
	}

	log.Fatal("Stream broke... dangit!")
}
	
