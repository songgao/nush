package nush

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/googollee/go-socket.io"
)

type HTTPListener struct {
	ch chan *Session
}

func NewHTTPListener(mux *http.ServeMux) (*HTTPListener, error) {
	httpListener := new(HTTPListener)
	httpListener.ch = make(chan *Session)
	server, err := socketio.NewServer(nil)
	if err != nil {
		return nil, err
	}
	server.On("connection", func(socket socketio.Socket) {
		httpListener.ch <- &Session{Stream: newSocketIOStream(socket), Context: fmt.Sprintf("http-%v-%s", socket.Request().URL, socket.Id())}
	})
	server.On("error", func(socket socketio.Socket, err error) {
		log.Println(err)
	})
	mux.Handle("/socket.io/", server)
	return httpListener, nil
}

func (l *HTTPListener) Accept() *Session {
	return <-l.ch
}

type socketIOStream struct {
	socket  socketio.Socket
	done    bool
	buf     chan string
	current *strings.Reader
}

func newSocketIOStream(socket socketio.Socket) *socketIOStream {
	stream := new(socketIOStream)
	stream.socket = socket
	stream.buf = make(chan string)
	socket.On("data", func(msg string) {
		stream.buf <- msg
	})
	socket.On("disconnection", func() {
		stream.done = true
	})
	return stream
}

func (s *socketIOStream) Read(p []byte) (int, error) {
	if s.done {
		return 0, io.ErrClosedPipe
	}

	if s.current == nil {
		// blocking call for the first one, so that Read always returns something
		s.current = strings.NewReader(<-s.buf)
	}

	var (
		n   int
		err error
	)
	var index int
	for index = 0; index < len(p); index += n {
		if s.current == nil {
			select {
			case str := <-s.buf:
				s.current = strings.NewReader(str)
			default:
			}
		}
		if s.current == nil {
			break
		}
		n, err = s.current.Read(p[index:])
		if err != nil {
			s.current = nil
		}
	}

	return index, nil
}

func (s *socketIOStream) Write(p []byte) (n int, err error) {
	if s.done {
		return 0, io.ErrClosedPipe
	}
	return len(p), s.socket.Emit("data", string(p))
}

func (s *socketIOStream) Close() error {
	s.done = true
	return s.socket.Emit("disconnection")
}
