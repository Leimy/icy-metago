package bot

import (
	"bufio"
	"fmt"
	"net"
	// "encoding/binary"
	// "crypto/rand"
	"log"
	"os"
	"regexp"
	"strings"

	"icy-metago/commands"
)

type Bot struct {
	room          string
	name          string
	serverAndPort string
	bior          *bufio.Reader
	biow          *bufio.Writer
	cc            chan commands.Command
	SChan         chan string
}

var userRegExp *regexp.Regexp
var actionRegExp *regexp.Regexp
var chanAndMessageRegExp *regexp.Regexp

func init() {
	// if this matches it produces string slices size 6
	// 0 is the whole string that matched
	// 1 is the nickname
	// 2 is the channel involved
	// 3 If it was an action, this is the string "ACTION " (trailing space intentional)
	// 4 Color
	// 5 The message delivered by the nick on the channel
	//	chanAndMessageRegExp = regexp.MustCompile("^:(.+)!.*PRIVMSG (#.+) :(ACTION )?(.+)$")
	chanAndMessageRegExp = regexp.MustCompile("^:([[:print:]]+)!.*PRIVMSG (#[[:print:]]+) :(ACTION )?[^[:digit:]]*?([[:print:]]+)$")
}

// Functions to be used externally to
// send commands to the bot
func (b *Bot) Quit() {
	b.cc <- &commands.QuitCmd{}
}

func (b *Bot) StringReplyCommand(s string, to ...string) {
	if len(to) == 0 {
		b.cc <- &commands.StringReplyCmd{nil, &s}
	} else {
		b.cc <- &commands.StringReplyCmd{&to[0], &s}
	}
}

func (b *Bot) SetInterval(i uint32) {
	b.cc <- &commands.SetIntervalCmd{&i}
}

// Interfaces we want to implement for Bot
func (b *Bot) Write(data []byte) (int, error) {
	return b.biow.Write(data)
}

func (b *Bot) Read(data []byte) (int, error) {
	return b.bior.Read(data)
}

func (b *Bot) ReadLine() ([]byte, bool, error) {
	return b.bior.ReadLine()
}

func (b *Bot) Flush() error {
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

// Gets data incoming from the IRC server
// aggregates it into one buffer and sends it to the channel
// forever
func (b *Bot) fromIRC(completeSChan chan<- string) {
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
	case commands.StringReply:
		fmt.Fprintf(b, "PRIVMSG %s :%s\r\n", b.room, *command.String())
		b.Flush()
	case commands.Quit:
		log.Printf("Shutting down because of a Quit command\n")
		os.Exit(0)
	case commands.SetInterval:
		log.Printf("Interval set req: %v\n", command.UInt32())
	}
}

func (b *Bot) parseTokens(lines []string) {
	if len(lines) == 0 {
		// this is ok, just pass
		return
	}
	if len(lines) < 5 {
		log.Panic(lines)
	}

	body := lines[4]

	log.Printf("4 == %q\n", body)

	if lines[3] == "ACTION " {
		// action handler
	} else if body == "?lastsong?" {
		// we pass this one to the shout server to get the metadata
		b.SChan <- body
	} else if body == "?quit?" {
		b.Quit()
	} else if strings.HasPrefix(body, b.name) {
		// We've been addressed... reply
		b.StringReplyCommand("Sup?", lines[1])
	}
	return
}

// process each line
func (b *Bot) procLine(line string) {
	log.Print(line)

	// Handle PING so we don't get hung up on.
	if strings.HasPrefix(line, "PING :") {
		resp := strings.Replace(line, "PING", "PONG", 1)
		fmt.Fprintf(b, "%s\r\n", resp)
	} else {
		b.parseTokens(chanAndMessageRegExp.FindStringSubmatch(line))
	}
	if err := b.Flush(); err != nil {
		log.Panic(err)
	}
}

// Just multiplex some channels for commands
func (b *Bot) loop() {
	completeSChan := make(chan string)
	go b.fromIRC(completeSChan)

	lchan := make(chan string)
	go func() {
		for line := range lchan {
			b.procLine(line)
		}
	}()

	for {
		select {
		case command := <-b.cc:
			b.procCommand(command)
		case line := <-completeSChan:
			lchan <- line
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
