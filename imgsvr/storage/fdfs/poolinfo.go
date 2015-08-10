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

func (this *poolInfo) makeConn() (net.Conn, error) {
	addr := fmt.Sprintf("%s:%d", this.host, this.port)
	conn, err := net.DialTimeout("tcp", addr, 30*time.Second)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (this *poolInfo) newPool() (pool.Pool, error) {
	p, err := pool.NewBlockingPool(this.minConns, this.maxConns, this.makeConn)
	if err != nil {
		return nil, err
	}
	return p, nil
}
