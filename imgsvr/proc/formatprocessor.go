package proc

import (
	l4g "github.com/alecthomas/log4go"
	"github.com/ctripcorp/cat"
	"github.com/ctripcorp/nephele/imgsvr/img4g"
)

type FormatProcessor struct {
	Format string
	Cat    cat.Cat
}

func (this *FormatProcessor) Process(img *img4g.Image) error {
	l4g.Debug("process format " + this.Format)
	var err error
	//tran := this.Cat.NewTransaction(Image, "Format")
	defer func() {
		//	tran.SetStatus(err)
		//	tran.Complete()
	}()
	err = img.SetFormat(this.Format)
	return err
}
