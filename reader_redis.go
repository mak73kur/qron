package qron

import (
	"bytes"
	"sync"
	"time"

	"github.com/mediocregopher/radix.v2/pool"
)

type RedisReader struct {
	p   *pool.Pool
	Key string

	sync.Mutex
	lt []byte
}

func newRedisPool(url, auth string, db int) (*pool.Pool, error) {
	p, err := pool.New("tcp", url, 1)
	if err != nil {
		return nil, err
	}
	if auth != "" {
		res := p.Cmd("AUTH", auth)
		if res.Err != nil {
			return nil, err
		}
	}
	if db != 0 {
		res := p.Cmd("SELECT", db)
		if res.Err != nil {
			return nil, err
		}
	}
	return p, nil
}

func NewRedisReader(url, auth string, db int) (*RedisReader, error) {
	p, err := newRedisPool(url, auth, db)
	if err != nil {
		return nil, err
	}
	return &RedisReader{p: p}, nil
}

func (r *RedisReader) Read() ([]byte, error) {
	tab, err := r.load()
	if err == nil {
		r.setLastTab(tab)
	}
	return tab, nil
}

// Implement Watcher interface
func (r *RedisReader) Watch(ch chan<- []byte) {
	for {
		time.Sleep(time.Minute)

		tab, err := r.load()
		if err != nil {
			continue
		}
		if !bytes.Equal(r.lastTab(), tab) {
			r.setLastTab(tab)
			ch <- tab
		}
	}
}

func (r *RedisReader) load() ([]byte, error) {
	res := r.p.Cmd("GET", r.Key)
	if res.Err != nil {
		return nil, res.Err
	}
	tab, err := res.Bytes()
	if err != nil {
		return nil, res.Err
	}
	return tab, nil
}

func (r *RedisReader) lastTab() []byte {
	r.Lock()
	defer r.Unlock()
	return r.lt
}

func (r *RedisReader) setLastTab(new []byte) {
	r.Lock()
	r.lt = new
	r.Unlock()
}
