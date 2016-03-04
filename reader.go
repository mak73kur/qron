package qron

type Reader interface {
	Read() ([]byte, error)
}

type Watcher interface {
	Watch(chan<- []byte)
}
