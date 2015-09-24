package proc

import (
	log "github.com/ctripcorp/nephele/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	cat "github.com/ctripcorp/nephele/Godeps/_workspace/src/github.com/ctripcorp/cat.go"
	"github.com/ctripcorp/nephele/imgsvr/img4g"
	"math"
)

type ResizeRProcessor struct {
	Width  int64
	Height int64
	Cat    cat.Cat
}

func (this *ResizeRProcessor) Process(img *img4g.Image) error {
	log.Debug("process resize r")
	var err error
	tran := this.Cat.NewTransaction("Command", "ResizeR")
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
	var (
		x int64 = 0
		y int64 = 0
		w int64 = width
		h int64 = height
	)
	if width > this.Width && height > this.Height {
		p1 := float64(this.Width) / float64(this.Height)
		p2 := float64(width) / float64(height)
		w = 0
		h = 0
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
	} else {
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

}
