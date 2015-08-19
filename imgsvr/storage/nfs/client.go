package nfs

import ( 
	"io/ioutil"
	"strings"
)

type NFSClient struct {
	Path string
}

type Error interface {
	Error() string
	Normal() bool //is normal error?
	Type() string //error type
}

type readError struct {
	error
}

func wrapError(err error) Error {
	return &readError{err}
}

func (e *readError) Type() string {
	if strings.Contains(e.Error(), "no such file or directory") {
		return "FileNotExistError"
	} else {
		return "UnExpectedError"
	}
}

func (e *readError) Normal() bool {
	if e.Type() == "FileNotExistError" {
		return true
	} else {
		return false
	}
}

func (this *NFSClient) Get() ([]byte, error) {
	buff, err := ioutil.ReadFile(this.Path)
	if err != nil {
		return nil, wrapError(err)
	} else {
		return buff, nil
	}
}
