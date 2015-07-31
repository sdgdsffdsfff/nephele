package proc

import (
	l4g "github.com/alecthomas/log4go"
	"github.com/ctripcorp/cat"
	"github.com/ctripcorp/nephele/imgsvr/img4g"
	"math"
)

type ResizeZProcessor struct {
	Width     int64
	Height    int64
	Cat       cat.Cat
	imgWidth  int64
	imgHeight int64
}

//高固定，宽（原图比例计算），宽固定，高（原图比例计算） （压缩）
func (this *ResizeZProcessor) Process(img *img4g.Image) error {
	l4g.Debug("process resize z")
	var err error
	tran := this.Cat.NewTransaction(Image, "ResizeW")
	defer func() {
		tran.SetStatus(err)
		tran.Complete()
	}()

	var width, height = this.imgWidth, this.imgHeight
	if this.imgWidth == 0 || this.imgHeight == 0 {
		var wd, ht int64
		wd, ht, err = img.Size()
		if err != nil {
			return err
		}
		width = wd
		height = ht
	}
	p1 := float64(this.Width) / float64(this.Height)
	p2 := float64(width) / float64(height)
	w, h := this.Width, this.Height
	if p2 > p1 {
		h = int64(math.Floor(float64(this.Width) / p2))
		if int64(math.Abs(float64(h-this.Height))) < 3 {
			h = this.Height
		}
	} else {
		w = int64(math.Floor(float64(this.Height) * p2))
		if int64(math.Abs(float64(w-this.Width))) < 3 {
			w = this.Width
		}
	}
	err = img.Resize(w, h)
	return err
}
