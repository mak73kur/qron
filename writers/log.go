package writers

import "log"

// Log is an example writer implementation that writes messages to program output
type Log struct{}

func (w Log) Write(msg []byte) error {
	log.Println(string(msg))
	return nil
}
