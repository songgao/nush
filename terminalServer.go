package nush

import (
	"fmt"
	"io"

	"golang.org/x/crypto/ssh/terminal"
)

type TerminalServer []Command

func (s TerminalServer) Serve(session *Session) {
	term := terminal.NewTerminal(session.Stream, "nush-0.0.1$ ")
	// there seems to be a bug here: the new ssh/terminal.Terminal inserts a \n
	// at end of each line. This sets terminal width to MAXINT so that it never
	// thinks it's end of line
	term.SetSize(int(^uint(0)>>1), 0)
	go func() {
		defer session.Stream.Close()
		term.Write(term.Escape.Red)
		io.WriteString(term, "This is still a WIP. For now it simply echos whatever you enter.\r\n\n")
		term.Write(term.Escape.Reset)
		for {
			line, err := term.ReadLine()
			if err != nil {
				break
			}
			logger.Printf("From %s: %s", session.Context, line)
			fmt.Fprintf(term, "Got: %s\n\r", line)
		}
	}()
}

func (s TerminalServer) ListenAndServe(listeners []Acceptor) {
	for _, listener := range listeners {
		go func(acceptor Acceptor) {
			for {
				s.Serve(acceptor.Accept())
			}
		}(listener)
	}
	select {}
}
