package qron

type Writer interface {
	Write(msg []byte) error
}
