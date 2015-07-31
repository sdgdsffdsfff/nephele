package proc

import (
	l4g "github.com/alecthomas/log4go"
	"github.com/ctripcorp/cat"
	"github.com/ctripcorp/nephele/imgsvr/img4g"
)

type ScaleProcessor struct {
	Width  int64
	Height int64
	Cat    cat.Cat
}

func (p *ScaleProcessor) Process(img *img4g.Image) error {
	l4g.Debug("process scale")
	var err error
	tran := cat.Instance().NewTransaction(Image, "Scale")
	defer func() {
		tran.SetStatus(err)
		tran.Complete()
	}()
	err = img.Resize(p.Width, p.Height)
	return err
}
