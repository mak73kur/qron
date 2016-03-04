package qron

import (
	"github.com/mediocregopher/radix.v2/pool"
)

type RedisWriter struct {
	p *pool.Pool

	Key      string
	LeftPush bool
}

func NewRedisWriter(url, auth string, db int) (*RedisWriter, error) {
	p, err := newRedisPool(url, auth, db)
	if err != nil {
		return nil, err
	}
	return &RedisWriter{p: p}, nil
}

func (r *RedisWriter) Write(msg []byte) error {
	var cmd string
	if r.LeftPush {
		cmd = "LPUSH"
	} else {
		cmd = "RPUSH"
	}
	res := r.p.Cmd(cmd, r.Key)
	return res.Err
}
