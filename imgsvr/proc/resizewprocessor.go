package proc

import (
	l4g "github.com/alecthomas/log4go"
	"github.com/ctripcorp/cat"
	"github.com/ctripcorp/nephele/imgsvr/img4g"
)

type ResizeWProcessor struct {
	Width  int64
	Height int64
	Cat    cat.Cat
}

//高固定，宽（原图比例计算），宽固定，高（原图比例计算） （压缩）
func (this *ResizeWProcessor) Process(img *img4g.Image) error {
	l4g.Debug("process resize w")
	var err error
	tran := this.Cat.NewTransaction(Image, "ResizeW")
	defer func() {
		tran.SetStatus(err)
		tran.Complete()
	}()
	var width, height int64
	width, height, err = img.Size()
	if err != nil {
		return err
	}
	if width <= this.Width && height <= this.Height {
		return nil
	}
	z := ResizeZProcessor{this.Width, this.Height, this.Cat, width, height}
	err = z.Process(img)
	return err
}
