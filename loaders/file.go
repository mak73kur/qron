package loaders

import "io/ioutil"

type File struct {
	path string
}

func NewFile(path string) File {
	return File{path}
}

func (l File) Load() (string, error) {
	data, err := ioutil.ReadFile(l.path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
