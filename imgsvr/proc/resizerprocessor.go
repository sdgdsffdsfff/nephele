package proc

import (
	l4g "github.com/alecthomas/log4go"
	"github.com/ctripcorp/cat"
	"github.com/ctripcorp/nephele/imgsvr/img4g"
)

type ResizeRProcessor struct {
	Width  int64
	Height int64
	Cat    cat.Cat
}

func (this *ResizeRProcessor) Process(img *img4g.Image) error {
	l4g.Debug("process resize r")
	var err error
	tran := this.Cat.NewTransaction(Image, "ResizeR")
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
	if width > this.Width && height > this.Height {
		c := ResizeCProcessor{this.Width, this.Height, this.Cat, width, height}
		err = c.Process(img)
		return err
	}
	var (
		x int64 = 0
		y int64 = 0
		w int64 = width
		h int64 = height
	)
	if width > this.Width {
		x = (w - this.Width) / 2
		w = this.Width
	}
	if height > this.Height {
		y = (h - this.Height) / 2
		h = this.Height
	}
	err = img.Crop(w, h, x, y)
	return err
}
