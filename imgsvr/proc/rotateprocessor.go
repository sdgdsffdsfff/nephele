package proc

import (
	log "github.com/ctripcorp/nephele/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	cat "github.com/ctripcorp/nephele/Godeps/_workspace/src/github.com/ctripcorp/cat.go"
	"github.com/ctripcorp/nephele/imgsvr/img4g"
)

type RotateProcessor struct {
	Degress float64
	Cat     cat.Cat
}

func (this *RotateProcessor) Process(img *img4g.Image) error {
	log.Debug("process rotate ")
	var err error
	tran := this.Cat.NewTransaction("Command", "Rotate")
	defer func() {
		tran.SetStatus(err)
		tran.Complete()
	}()
	err = img.Rotate(this.Degress)
	return err
}
