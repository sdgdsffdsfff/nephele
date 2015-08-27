package proc

import (
	log "github.com/ctripcorp/nephele/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	cat "github.com/ctripcorp/nephele/Godeps/_workspace/src/github.com/ctripcorp/cat.go"
	"github.com/ctripcorp/nephele/imgsvr/img4g"
)

type StripProcessor struct {
	Cat cat.Cat
}

func (this *StripProcessor) Process(img *img4g.Image) error {
	log.Debug("process strip")
	err := img.Strip()
	return err
}
