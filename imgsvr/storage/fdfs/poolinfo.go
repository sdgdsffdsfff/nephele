package fdfs

import (
	"fmt"
	"github.com/ctripcorp/cat"
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
	Cat := cat.Instance()
	addr := fmt.Sprintf("%s:%d", this.host, this.port)
	conn, err := net.DialTimeout("tcp", addr, 30*time.Second)
	if err != nil {
		event := Cat.NewEvent("DialFdfs", "Fail")
		event.AddData("addr", addr)
		event.AddData("detail", err.Error())
		event.SetStatus("ERROR")
		event.Complete()
		return nil, err
	}
	event := Cat.NewEvent("DialFdfs", "Success")
	event.AddData("addr", addr)
	event.SetStatus("0")
	event.Complete()
	return conn, nil
}

func (this *poolInfo) newPool() (pool.Pool, error) {
	p, err := pool.NewBlockingPool(this.minConns, this.maxConns, this.makeConn)
	if err != nil {
		return nil, err
	}
	return p, nil
}
