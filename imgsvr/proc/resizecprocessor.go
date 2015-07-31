package proc

import (
	l4g "github.com/alecthomas/log4go"
	"github.com/ctripcorp/cat"
	"github.com/ctripcorp/nephele/imgsvr/img4g"
	"math"
)

type ResizeCProcessor struct {
	Width     int64
	Height    int64
	Cat       cat.Cat
	imgWidth  int64
	imgHeight int64
}

func (this *ResizeCProcessor) Process(img *img4g.Image) error {
	l4g.Debug("process resize c")
	var err error
	tran := this.Cat.NewTransaction(Image, "ResizeC")
	defer func() {
		tran.SetStatus(err)
		tran.Complete()
	}()

	var width, height = this.imgWidth, this.imgHeight
	var wd, ht int64
	if width == 0 || height == 0 {
		wd, ht, err = img.Size()
		if err != nil {
			return err
		}
		width = wd
		height = ht
	}

	p1 := float64(this.Width) / float64(this.Height)
	p2 := float64(width) / float64(height)
	var (
		x int64 = 0
		y int64 = 0
		w int64 = 0
		h int64 = 0
	)
	if math.Abs(p1-p2) > 0.0001 {
		if p2 > p1 { //以高缩小
			h = height
			w = int64(math.Floor(float64(h) * p1))
			x = (width - w) / 2
		}
		if p2 < p1 { //以宽缩小
			w = width
			h = int64(math.Floor(float64(w) / p1))
			y = (height - h) / 2
		}
		err = img.Crop(w, h, x, y)
		if err != nil {
			return err
		}
	}
	err = img.Resize(this.Width, this.Height)
	return err
}
