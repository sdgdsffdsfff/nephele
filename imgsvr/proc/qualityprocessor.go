package proc

import (
	l4g "github.com/alecthomas/log4go"
	"github.com/ctripcorp/cat"
	"github.com/ctripcorp/nephele/imgsvr/img4g"
)

type QualityProcessor struct {
	Quality int
	Cat     cat.Cat
}

func (this *QualityProcessor) Process(img *img4g.Image) error {
	l4g.Debug("process quality ")
	var err error
	tran := this.Cat.NewTransaction(Image, "Quality")
	defer func() {
		tran.SetStatus(err)
		tran.Complete()
	}()
	err = img.SetCompressionQuality(this.Quality)
	return err
}
