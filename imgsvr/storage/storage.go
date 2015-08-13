package storage

import (
	"github.com/ctripcorp/cat"
	"github.com/ctripcorp/nephele/imgsvr/storage/fdfs"
	"github.com/ctripcorp/nephele/imgsvr/storage/nfs"
	"strconv"
)

type Storage interface {
	GetImage() ([]byte, error)
}

type Fdfs struct {
	Path          string
	TrackerDomain string
	Port          int
	Cat           cat.Cat
}

var client fdfs.FdfsClient = nil
var lock chan int = make(chan int, 1)
var initialized bool = false

func (this *Fdfs) GetImage() ([]byte, error) {
	if client == nil {
		lock <- 0
		if !initialized {
			if client == nil {
				var e error
				client, e = fdfs.NewFdfsClient([]string{this.TrackerDomain}, strconv.Itoa(this.Port))
				if e != nil {
					return nil, e
				}
			}
			initialized = true
		}
		<-lock
	}

	bts, err := client.DownloadToBuffer(this.Path)
	if err != nil {
		return nil, err
	} else {
		return bts, nil
	}
}

type Nfs struct {
	Path string
}

func (f *Nfs) GetImage() ([]byte, error) {
	n := &nfs.NFSClient{f.Path}
	return n.Get()
}
