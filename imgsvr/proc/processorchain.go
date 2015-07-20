package proc

import (
	"errors"
	"github.com/ctripcorp/nephele/imgsvr/img4g"
)

type ImageProcessor interface {
	Process(*img4g.Image) error
}

type ProcessorChain struct {
	Chain []ImageProcessor
}

func (p *ProcessorChain) Process(img *img4g.Image) error {
	if len(p.Chain) == 0 {
		return errors.New("procchain.unexpected.mark(len:0)")
	}

	for _, proc := range p.Chain {
		err := proc.Process(img)
		if err != nil {
			return err
		}
	}

	return nil
}
