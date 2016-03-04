package qron

type InlineReader struct {
	Tab []byte
}

func (l InlineReader) Read() ([]byte, error) {
	return l.Tab, nil
}
