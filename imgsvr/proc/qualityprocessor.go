package proc

import (
	log "github.com/ctripcorp/nephele/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	cat "github.com/ctripcorp/nephele/Godeps/_workspace/src/github.com/ctripcorp/cat.go"
	"github.com/ctripcorp/nephele/imgsvr/img4g"
)

type QualityProcessor struct {
	Quality int
	Cat     cat.Cat
}

func (this *QualityProcessor) Process(img *img4g.Image) error {
	log.WithFields(log.Fields{
		"quality": this.Quality,
	}).Debug("process quality")
	err := img.SetCompressionQuality(this.Quality)
	return err
}
