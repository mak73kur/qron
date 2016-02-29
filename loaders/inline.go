package loaders

type Inline struct {
	tab string
}

func NewInline(tab string) Inline {
	return Inline{tab}
}

func (l Inline) Load() (string, error) {
	return l.tab, nil
}
