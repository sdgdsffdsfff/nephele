package nfs

import "io/ioutil"

type NFSClient struct {
	Path string
}

func (this *NFSClient) Get() ([]byte, error) {
	return ioutil.ReadFile(this.Path)
}
