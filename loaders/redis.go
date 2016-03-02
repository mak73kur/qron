package loaders

import (
	"sync"
	"time"

	"github.com/mediocregopher/radix.v2/pool"
)

type Redis struct {
	Key string
	p   *pool.Pool

	sync.Mutex
	lt string
}

func NewRedis(url, key string) (*Redis, error) {
	p, err := pool.New("tcp", url, 1)
	if err != nil {
		return nil, err
	}
	return &Redis{p: p, Key: key}, nil
}

func (r *Redis) Select(db int) error {
	res := r.p.Cmd("SELECT", db)
	return res.Err
}

func (r *Redis) Auth(pass string) error {
	res := r.p.Cmd("AUTH", pass)
	return res.Err
}

func (r *Redis) Load() (string, error) {
	str, err := r.Load()
	if err == nil {
		r.setLastTab(str)
	}
	return str, nil
}

func (r *Redis) load() (string, error) {
	res := r.p.Cmd("GET", r.Key)
	if res.Err != nil {
		return "", res.Err
	}
	str, err := res.Str()
	if err != nil {
		return "", res.Err
	}
	return str, nil
}

func (r *Redis) Poll(ch chan<- string, errCh chan<- error) {
	for {
		time.Sleep(time.Minute)

		str, err := r.load()
		if err != nil {
			errCh <- err
			continue
		}

		if r.lastTab() != str {
			r.setLastTab(str)
			ch <- str
		}
	}
}

func (r *Redis) lastTab() string {
	r.Lock()
	defer r.Unlock()
	return r.lt
}

func (r *Redis) setLastTab(new string) {
	r.Lock()
	r.lt = new
	r.Unlock()
}
