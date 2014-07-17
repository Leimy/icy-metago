package bot

import (
	"net"
	"bufio"
	"fmt"
	// "encoding/binary"
	// "crypto/rand"
	// "regexp"
	"os"
	"strings"
	"log"

	"icy-metago/commands"
)

type Bot struct {
	room string
	name string
	serverAndPort string
	bior *bufio.Reader
	biow *bufio.Writer
	cc chan commands.Command
	SChan chan string
}
// Functions to be used externally to
// send commands to the bot
func (b *Bot) Quit () {
	b.cc <- &commands.QuitCmd{}
}

func (b *Bot) SetMeta (s string) {
	b.cc <- &commands.SetMetaCmd{&s}
}

func (b *Bot) SetInterval (i uint32) {
	b.cc <- &commands.SetIntervalCmd{&i}
}


// Interfaces we want to implement for Bot
func (b *Bot) Write (data []byte) (int, error) {
	return b.biow.Write(data)
}

func (b *Bot) Read (data []byte) (int, error) {
	return b.bior.Read(data)
}

func (b *Bot) ReadLine () ([]byte, bool, error) {
	return b.bior.ReadLine()
}

func (b *Bot) Flush () error {
	return b.biow.Flush()
}

// Just some setup stuff for getting into the channel
func (b *Bot) loginstuff() {
	fmt.Fprintf(b, "NICK %s\r\n", b.name)
	fmt.Fprintf(b, "USER %s 0 * :tutorial bot\r\n", b.name)
	fmt.Fprintf(b, "JOIN %s\r\n", b.room)
	if err := b.Flush(); err != nil {
		log.Panic(err)
	}
}

// Gets data from the bot, incoming from the IRC server
// aggregates it into one buffer and sends it to the channel
// forever
func (b *Bot) fromIRC(completeSChan chan <-string) {
	for {
		bytes, morep, err := b.ReadLine()
		if err != nil {
			log.Panic(err)
		}
		for morep {
			bytes2, morep, err := b.ReadLine()
			if err != nil {
				log.Panic(err)
			}
			bytes = append(bytes, bytes2...)
			if !morep {
				break
			}
		}
		completeSChan <- string(bytes)
	}
}

// This is where command behaviors are implemented
func (b *Bot) procCommand(command commands.Command) {
	switch command.Code() {
	case commands.SetMeta:
		fmt.Fprintf(b, "PRIVMSG %s :%s\r\n", b.room, *command.String())
		b.Flush()
	case commands.Quit:
		log.Printf("Shutting down because of a Quit command\n")
		os.Exit(0)
	case commands.SetInterval:
		log.Printf("Interval set req: %v\n", command.UInt32())
	}
}

func (b *Bot) procLine(line string) {
	log.Print(line)

	// Handle PING so we don't get hung up on.
	if strings.HasPrefix(line, "PING :") {
		resp := strings.Replace(line, "PING", "PONG", 1)
		fmt.Fprintf(b, "%s\r\n", resp)
	} else {
		b.SChan <- line  // TODO: make this (string, string) for nick, text
	}
}

func (b *Bot) loop () {
	completeSChan := make(chan string)
	go b.fromIRC(completeSChan) 
	
	for {
		select {
		case command := <- b.cc:
			b.procCommand(command)
		case line := <- completeSChan:
			b.procLine(line)
			if err := b.Flush(); err != nil { log.Panic(err) }
		}
	}
}

func bot(room, name, serverAndport string) (*Bot, error) {
	log.Printf("IRC bot connecting to %s as %s to channel %s\n",
		serverAndport, name, room)
	conn, err := net.Dial("tcp4", serverAndport)
	if err != nil {
		return nil, err
	}
	log.Print("Done connecting")

	bot := &Bot{
		room, 
		name, 
		serverAndport, 
		bufio.NewReader(conn), 
		bufio.NewWriter(conn),
		make(chan commands.Command),
		make(chan string)}

	bot.loginstuff()
	go bot.loop()

	return bot, nil
}

func NewBot(room, name, serverAndPort string) (*Bot, error) {
	return bot(room, name, serverAndPort)
}
