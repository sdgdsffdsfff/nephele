package proc

import (
	log "github.com/ctripcorp/nephele/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	cat "github.com/ctripcorp/nephele/Godeps/_workspace/src/github.com/ctripcorp/cat.go"
	"github.com/ctripcorp/nephele/imgsvr/img4g"
)

type FormatProcessor struct {
	Format string
	Cat    cat.Cat
}

func (this *FormatProcessor) Process(img *img4g.Image) error {
	log.WithFields(log.Fields{
		"format": this.Format,
	}).Debug("process format")
	err := img.SetFormat(this.Format)
	return err
}
