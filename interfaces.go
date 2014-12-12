package nush

import "io"

type Command interface {
}

type Acceptor interface {
	Accept() io.ReadWriteCloser
}
