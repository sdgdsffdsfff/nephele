package proc

import (
	l4g "github.com/alecthomas/log4go"
	"github.com/ctripcorp/cat"
	"github.com/ctripcorp/nephele/imgsvr/img4g"
)

type RotateProcessor struct {
	Degress float64
	Cat     cat.Cat
}

func (this *RotateProcessor) Process(img *img4g.Image) error {
	l4g.Debug("process rotate ")
	var err error
	tran := this.Cat.NewTransaction(Image, "Rotate")
	defer func() {
		tran.SetStatus(err)
		tran.Complete()
	}()
	err = img.Rotate(this.Degress)
	return err
}
