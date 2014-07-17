package commands

type CommandCode int

const (
	SetMeta CommandCode = iota
	SetInterval
	Quit
)

type Command interface {
	Code() CommandCode
	String() *string
	UInt32() *uint32
}

type SetMetaCmd struct {
	S *string
}

func (mc *SetMetaCmd) Code() CommandCode { return SetMeta }
func (mc *SetMetaCmd) String() *string   { return mc.S }
func (mc *SetMetaCmd) UInt32() *uint32   { return nil }

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
