package loaders

import "io/ioutil"

type File struct {
	Path string
}

func (l File) Load() (string, error) {
	data, err := ioutil.ReadFile(l.Path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
