package qron

type Writer interface {
	Write([]byte) error
}
