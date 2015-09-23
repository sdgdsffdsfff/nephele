package nfs

import (
	"io/ioutil"
	"net/http"
)

type NFSClient struct {
	Path string
}

type FileNotExistError string

func (e FileNotExistError) Type() string { return "FileNotExistError" }

func (e FileNotExistError) Normal() bool { return true }

func (e FileNotExistError) Error() string { return "no such file or directory:" + string(e) }

func (this *NFSClient) Get() ([]byte, error) {
	resp, err := http.Get(this.Path)
	if err != nil {
		// handle error
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return nil, FileNotExistError(this.Path)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	} else {
		return body, nil
	}
}
