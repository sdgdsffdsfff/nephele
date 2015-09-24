package nfs

import (
	"io/ioutil"
	"net/http"
	"strings"
	"strconv"
)

type NFSClient struct {
	Path string
}

type FileNotExistError string

func (e FileNotExistError) Type() string { return "FileNotExistError" }

func (e FileNotExistError) Normal() bool { return true }

func (e FileNotExistError) Error() string { return "no such file or directory:" + string(e) }

type HttpStatusError struct {
	path       string
	statusCode int
}

func (e *HttpStatusError) Type() string { return "HttpStatusError" }

func (e *HttpStatusError) Normal() bool { return true }

func (e *HttpStatusError) Error() string {
	return "http status error: " + strconv.Itoa(e.statusCode) + ", request path: " + string(e.path)
}

func (this *NFSClient) Get() (b []byte, e error) {
	if strings.Contains(this.Path, "http://") {
		b, e = this.httpGet(this.Path)
	} else {
		b, e = this.localGet(this.Path)
	}
	return 
}

func (this *NFSClient) httpGet(path string) ([]byte, error) {
	resp, err := http.Get(path)
	if err != nil {
		// handle error
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return nil, FileNotExistError(this.Path)
	}
	if resp.StatusCode != 200 {
		return nil, &HttpStatusError{path, resp.StatusCode}
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	} else {
		return body, nil
	}
}

func (this *NFSClient) localGet(path string) ([]byte, error) {
	buff, err := ioutil.ReadFile(path)
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			return nil, FileNotExistError(path)
		} else {
			return nil, err
		}
	}
	return buff, nil
}
