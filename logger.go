package nush

import (
	"io"
	"io/ioutil"
	"log"
)

var logger = log.New(ioutil.Discard, "nush: ", log.Lshortfile)

func SetLoggerOutput(w io.Writer) {
	logger.SetOutput(w)
}
