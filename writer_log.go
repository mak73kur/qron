package qron

import "log"

// Log is an example writer implementation that writes messages to program output
type LogWriter struct{}

func (w LogWriter) Write(msg []byte) error {
	log.Println(string(msg))
	return nil
}
