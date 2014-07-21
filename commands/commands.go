package commands

import "fmt"

type CommandCode int

const (
	StringReply CommandCode = iota
	SetInterval
	Quit
)

type Command interface {
	Code() CommandCode
	String() *string
	UInt32() *uint32
}

type StringReplyCmd struct {
	To   *string
	Mess *string
}

func (mc *StringReplyCmd) Code() CommandCode { return StringReply }
func (mc *StringReplyCmd) String() *string {
	if mc.To != nil {
		s := fmt.Sprintf("%s: %s", *mc.To, *mc.Mess)
		return &s
	} else {
		return mc.Mess
	}
}
func (mc *StringReplyCmd) UInt32() *uint32 { return nil }

type SetIntervalCmd struct {
	I *uint32
}

func (si *SetIntervalCmd) Code() CommandCode { return SetInterval }
func (si *SetIntervalCmd) String() *string   { return nil }
func (si *SetIntervalCmd) UInt32() *uint32   { return si.I }

type QuitCmd struct {
	b byte
}

func (q *QuitCmd) Code() CommandCode { return Quit }
func (q *QuitCmd) String() *string   { return nil }
func (q *QuitCmd) UInt32() *uint32   { return nil }
