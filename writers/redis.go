package writers

import (
	"github.com/mediocregopher/radix.v2/pool"
)

type Redis struct {
	p *pool.Pool

	Key      string
	LeftPush bool
}

func NewRedis(url string) (*Redis, error) {
	p, err := pool.New("tcp", url, 1)
	if err != nil {
		return nil, err
	}
	return &Redis{p: p}, nil
}

func (r *Redis) Select(db int) error {
	res := r.p.Cmd("SELECT", db)
	return res.Err
}

func (r *Redis) Auth(pass string) error {
	res := r.p.Cmd("AUTH", pass)
	return res.Err
}

func (r *Redis) Write(msg []byte) error {
	var cmd string
	if r.LeftPush {
		cmd = "LPUSH"
	} else {
		cmd = "RPUSH"
	}
	res := r.p.Cmd(cmd, r.Key)
	return res.Err
}
