package qron

// Qron tab loader
type Loader interface {
	Load() (string, error)
}

type Poller interface {
	Poll(chan<- string, chan<- error)
}
