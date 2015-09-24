package proc

import (
	log "github.com/ctripcorp/nephele/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	cat "github.com/ctripcorp/nephele/Godeps/_workspace/src/github.com/ctripcorp/cat.go"
	"github.com/ctripcorp/nephele/imgsvr/img4g"
)

type ScaleProcessor struct {
	Width  int64
	Height int64
	Cat    cat.Cat
}

func (p *ScaleProcessor) Process(img *img4g.Image) error {
	log.Debug("process scale")
	var err error
	tran := cat.Instance().NewTransaction("Command", "Scale")
	defer func() {
		tran.SetStatus(err)
		tran.Complete()
	}()
	err = img.Resize(p.Width, p.Height)
	return err
}
