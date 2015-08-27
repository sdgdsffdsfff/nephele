package proc

import (
	"errors"
	log "github.com/ctripcorp/nephele/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	cat "github.com/ctripcorp/nephele/Godeps/_workspace/src/github.com/ctripcorp/cat.go"
	"github.com/ctripcorp/nephele/imgsvr/img4g"
)

type WaterMarkProcessor struct {
	Logo          *img4g.Image
	Location      int
	Dissolve      int
	Cat           cat.Cat
	WaterMarkType string
}

func (this *WaterMarkProcessor) Process(img *img4g.Image) error {
	log.Debug("process watermark")
	var err error = nil
	tran := this.Cat.NewTransaction("Command", this.WaterMarkType)

	defer func() {
		this.Logo.DestoryWand()
		tran.SetStatus(err)
		tran.Complete()
	}()
	this.Logo.CreateWand()
	if this.Location == 0 {
		this.Location = 9
	}
	if this.Location < 1 || this.Location > 9 {
		err = errors.New("Logo location(" + string(this.Location) + ") isn't right!")
		return err
	}
	var x, y int64
	x, y, err = getLocation(this.Location, img, this.Logo)
	if err != nil {
		return err
	}

	if this.Dissolve > 0 && this.Dissolve < 100 {
		this.Logo.Dissolve(this.Dissolve)
	}

	err = img.Composite(this.Logo, x, y)
	return err
}

func getLocation(location int, img *img4g.Image, logo *img4g.Image) (int64, int64, error) {
	var (
		x int64 = 0
		y int64 = 0
	)
	width, height, err := img.Size()
	if err != nil {
		return 0, 0, err
	}
	logowidth, logoheight, err := logo.Size()
	if err != nil {
		return 0, 0, err
	}
	switch location {
	case 1:
		x, y = 0, 0
	case 2:
		x, y = (width-logowidth)/2, 0
	case 3:
		x, y = width-logowidth, 0
	case 4:
		x, y = 0, (height-logoheight)/2
	case 5:
		x, y = (width-logowidth)/2, (height-logoheight)/2
	case 6:
		x, y = width-logowidth, (height-logoheight)/2
	case 7:
		x, y = 0, height-logoheight
	case 8:
		x, y = (width-logowidth)/2, height-logoheight
	case 9:
		x, y = width-logowidth, height-logoheight
	}
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	return x, y, nil
}
