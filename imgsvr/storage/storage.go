package storage

import (
	"github.com/ctripcorp/nephele/imgsvr/storage/fdfs"
	"github.com/ctripcorp/nephele/imgsvr/storage/nfs"
)

type Storage interface {
	GetImage() ([]byte, error)
}

type Fdfs struct {
	Path          string
	TrackerDomain string
	Port          int
}

func (this *Fdfs) GetImage() ([]byte, error) {
	tracker := &fdfs.Tracker{
		[]string{this.TrackerDomain},
		this.Port,
	}

	fdfs, err := fdfs.NewFdfsClientByTracker(tracker)
	if err != nil {
		return nil, err
	} else {
		reponse, err := fdfs.DownloadToBuffer(this.Path)
		if err != nil {
			return nil, err
		} else {
			bts := reponse.Content.([]byte)
			return bts, nil
		}
	}
}

type Nfs struct {
	Path string
}

func (f *Nfs) GetImage() ([]byte, error) {
	n := &nfs.NFSClient{f.Path}
	return n.Get()
}
