package nush

import (
	"io"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

type SSHListener struct {
	listener net.Listener
	config   *ssh.ServerConfig
	ch       chan ssh.Channel
}

func NewSSHListener(config *ssh.ServerConfig, laddr string) (*SSHListener, error) {
	var err error
	l := new(SSHListener)
	l.config = config
	l.ch = make(chan ssh.Channel)
	l.listener, err = net.Listen("tcp", laddr)
	go l.accept()
	return l, err
}

func (l *SSHListener) accept() {
	for {
		conn, err := l.listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		_, chans, reqs, err := ssh.NewServerConn(conn, l.config)
		if err != nil {
			continue
		}
		go ssh.DiscardRequests(reqs)
		go func(chans <-chan ssh.NewChannel) {
			for channel := range chans {
				if channel.ChannelType() != "session" {
					channel.Reject(ssh.UnknownChannelType, "unknown channel type")
					continue
				}
				c, r, err := channel.Accept()
				if err != nil {
					continue
				}
				go func(requests <-chan *ssh.Request) {
					for req := range requests {
						if (req.Type == "shell" && len(req.Payload) == 0) || req.Type == "pty-req" {
							req.Reply(true, nil)
						} else {
							req.Reply(false, nil)
						}
					}
				}(r)
				l.ch <- c
			}
		}(chans)
	}
}

func (l *SSHListener) Accept() io.ReadWriteCloser {
	return <-l.ch
}
