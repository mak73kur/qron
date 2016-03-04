package qron

import "io/ioutil"

type FileReader struct {
	Path string
}

func (f FileReader) Read() ([]byte, error) {
	return ioutil.ReadFile(f.Path)
}
