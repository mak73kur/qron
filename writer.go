package qron

type Writer interface {
	Write([]byte, map[string]interface{}) error
}
