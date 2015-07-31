package proc

import (
	l4g "github.com/alecthomas/log4go"
	"github.com/ctripcorp/cat"
	"github.com/ctripcorp/nephele/imgsvr/img4g"
)

type StripProcessor struct {
	Cat cat.Cat
}

func (this *StripProcessor) Process(img *img4g.Image) error {
	l4g.Debug("process strip")
	var err error
	//tran := this.Cat.NewTransaction(Image, "Strip")
	defer func() {
		//	tran.SetStatus(err)
		//	tran.Complete()
	}()
	err = img.Strip()
	return err
}
