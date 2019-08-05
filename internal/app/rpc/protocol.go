package rpc

type ServerResponse string
type ClientCommand string

const (
	ServerOK  ServerResponse = "OK"
	ServerErr                = "ERR"
)

const (
	ClientSetMode ClientCommand = "SETMODE"
	ClientStop    ClientCommand = "STOP"
)

