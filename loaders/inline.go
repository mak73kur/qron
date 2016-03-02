package loaders

type Inline struct {
	Tab string
}

func (l Inline) Load() (string, error) {
	return l.Tab, nil
}
