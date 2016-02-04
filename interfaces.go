package nush

import "io"

type Command interface {
}

type Session struct {
	Stream  io.ReadWriteCloser
	Context string
}

type Acceptor interface {
	Accept() *Session
}
