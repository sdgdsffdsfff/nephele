package fdfs

import (
	"fmt"
	"github.com/ctripcorp/ghost/pool"
	"net"
	"time"
)

type poolInfo struct {
	host     string
	port     int
	minConns int
	maxConns int
}

func (this *poolInfo) makeConn() (conn net.Conn, err error) {
	addr := fmt.Sprintf("%s:%d", this.host, this.port)
	//try two times
	for i := 0; i < 2; i++ {
		if conn, err = net.DialTimeout("tcp", addr, time.Minute); err == nil {
			return
		}
	}
	return
}

func (this *poolInfo) newPool() (pool.Pool, error) {
	p, err := pool.NewBlockingPool(this.minConns, this.maxConns, this.makeConn)
	if err != nil {
		return nil, err
	}
	return p, nil
}
