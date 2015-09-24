package imgsvr

import (
	"errors"
	log "github.com/ctripcorp/nephele/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	cat "github.com/ctripcorp/nephele/Godeps/_workspace/src/github.com/ctripcorp/cat.go"
	"github.com/ctripcorp/nephele/imgsvr/data"
	"github.com/ctripcorp/nephele/imgsvr/img4g"
	"github.com/ctripcorp/nephele/imgsvr/proc"
	"strconv"
	"strings"
)

type ProcChainBuilder struct {
	Cat cat.Cat
}

var (
	CmdStrip     = "s"
	CmdFormat    = "f"
	CmdResize    = "resize"
	CmdQuality   = "q"
	CmdWaterMark = "m"
	CmdRotate    = "rotate"
)

type buildError struct {
	error
	errType string
}

func (e *buildError) Type() string {
	return e.errType
}

func (this *ProcChainBuilder) Build(params map[string]string) (*proc.ProcessorChain, *buildError) {
	procChain := &proc.ProcessorChain{Chain: make([]proc.ImageProcessor, 0, 5)}

	sourceType, channel, path := ParseUri(params[":1"])
	//_, channel, _ := ParseUri(params[":1"])
	sequences, e := data.GetSequenceofoperation(channel)
	if e != nil {
		return nil, &buildError{e, "UrlSequenceCmdError"}
	}
	for _, t := range sequences {
		switch t {
		case CmdStrip:
			stripProcessor, e := this.getStripProcessor(channel, params)
			if e != nil {
				return nil, &buildError{e, "UrlStripCmdError"}
			}
			procChain.Chain = append(procChain.Chain, stripProcessor)
			log.Debug("add strip processor")
		case CmdResize:
			resizeProcessor, e := this.getResizeProcessor(channel, params)
			if e != nil {
				return nil, &buildError{e, "UrlResizeCmdError"}
			}
			procChain.Chain = append(procChain.Chain, resizeProcessor)
			log.Debug("add resize processor")
		case CmdQuality:
			qualityProcessor, e := this.getQualityProcessor(channel, params)
			if e != nil {
				return nil, &buildError{e, "UrlQualityCmdError"}
			}
			if qualityProcessor != nil {
				procChain.Chain = append(procChain.Chain, qualityProcessor)
				log.Debug("add quality processor")
			}
		case CmdRotate:
			rotateProcessor, e := this.getRotateProcessor(channel, params)
			if e != nil {
				return nil, &buildError{e, "UrlRotateCmdError"}
			}
			if rotateProcessor != nil {
				procChain.Chain = append(procChain.Chain, rotateProcessor)
				log.Debug("add rotate processor")
			}
		case CmdWaterMark:
			waterMarkProcessors, e := this.getWaterMarkProcessors(sourceType, channel, path, params)
			if e != nil {
				return nil, &buildError{e, "UrlWaterMarkCmdError"}
			}
			if waterMarkProcessors != nil {
				for _, p := range waterMarkProcessors {
					if p != nil {
						procChain.Chain = append(procChain.Chain, p)
					}
				}
			}
		case CmdFormat:
			formatProcessor, e := this.getFormatProcessor(channel, params)
			if e != nil {
				return nil, &buildError{e, "UrlFormatCmdError"}
			}
			if formatProcessor != nil {
				procChain.Chain = append(procChain.Chain, formatProcessor)
				log.Debug("add format processor")
			}
		}
	}
	return procChain, nil
}

func (this *ProcChainBuilder) getFormatProcessor(channel string, params map[string]string) (proc.ImageProcessor, error) {
	ext, _ := params["ext"]
	return &proc.FormatProcessor{ext, this.Cat}, nil
}

func (this *ProcChainBuilder) getStripProcessor(channel string, params map[string]string) (proc.ImageProcessor, error) {
	return &proc.StripProcessor{this.Cat}, nil
}

func (this *ProcChainBuilder) getResizeProcessor(channel string, params map[string]string) (proc.ImageProcessor, error) {
	cmdVal, ok := params[":2"]
	if !ok {
		return nil, errors.New("proc.command.notfound.mark()")
	}
	widthVal, ok := params[":3"]
	if !ok {
		return nil, errors.New("image.width.notfound.mark()")
	}
	heightVal, ok := params[":4"]
	if !ok {
		return nil, errors.New("image.height.notfound.mark()")
	}
	cmd := strings.ToLower(cmdVal)
	width, height, err := this.getValidSizeParam(widthVal, heightVal, cmd, channel)
	if err != nil {
		return nil, err
	}

	//feature
	var (
		process proc.ImageProcessor = nil
		isnext  bool                = true
	)
	switch {
	case channel == TG:
		ft := tgresizefeature{width, height, cmd, this.Cat}
		process, isnext, err = ft.Process()
	case channel == Hotel || channel == Globalhotel:
		ft := hotelresizefeature{width, height, cmd, this.Cat}
		process, isnext, err = ft.Process()
	}
	if err != nil {
		return nil, err
	}
	if process != nil {
		return process, nil
	}
	if !isnext {
		return nil, nil
	}

	switch cmd {
	case "r":
		return &proc.ResizeRProcessor{width, height, this.Cat}, nil
	case "c":
		return &proc.ResizeCProcessor{Width: width, Height: height, Cat: this.Cat}, nil
	case "w":
		return &proc.ResizeWProcessor{width, height, this.Cat}, nil
	case "z":
		return &proc.ResizeZProcessor{Width: width, Height: height, Cat: this.Cat}, nil
	}
	return nil, nil
}

func (this *ProcChainBuilder) getValidSizeParam(widthVal, heightVal, cmdVal, channel string) (int64, int64, error) {
	width, err := strconv.ParseInt(widthVal, 10, 64)
	if err != nil {
		return 0, 0, err
	}
	height, err := strconv.ParseInt(heightVal, 10, 64)
	if err != nil {
		return 0, 0, err
	}

	//check type
	resizetypes, err := data.GetResizeTypes(channel)
	if err != nil {
		return 0, 0, err
	}

	if !strings.Contains(resizetypes, cmdVal) {
		return 0, 0, errors.New(JoinString("channel: ", channel, ", reason: not support type ", cmdVal))
	}

	//check size
	sizes, err := data.GetSizes(channel)
	if err != nil {
		return 0, 0, err
	}
	var wh = JoinString(",", widthVal, "x", heightVal, ",")
	if !strings.Contains(sizes, wh) {
		return 0, 0, errors.New(JoinString("channel: ", channel, ", reason: not support size ", wh))
	}
	return width, height, nil
}

func (this *ProcChainBuilder) getRotateProcessor(channel string, params map[string]string) (proc.ImageProcessor, error) {
	rotate, ok := params[":6"]
	if !ok {
		return nil, nil
	}
	degress, err := strconv.ParseFloat(rotate, 64)
	if err != nil {
		return nil, err
	}

	var (
		process proc.ImageProcessor = nil
		isnext  bool                = true
	)
	if channel == Hotel || channel == Globalhotel {
		ft := hotelrotatefeature{degress}
		process, isnext, err = ft.Process()
	}
	if err != nil {
		return nil, err
	}
	if process != nil {
		return process, nil
	}
	if !isnext {
		return nil, nil
	}

	const key = "rotates"
	rotateStr, err := data.GetRotates(channel)
	if err != nil {
		return nil, err
	}
	if !strings.Contains(rotateStr, JoinString(",", rotate, ",")) {
		return nil, errors.New(JoinString("channel: ", channel, ", reason: not support rotate degree ", rotate))
	}

	return &proc.RotateProcessor{degress, this.Cat}, nil
}

func (this *ProcChainBuilder) getQualityProcessor(channel string, params map[string]string) (proc.ImageProcessor, error) {
	var (
		quality int
		err     error
	)
	qualityStr, _ := params["n0"]
	if qualityStr != "" {
		qualitiesStr, err := data.GetQualities(channel)
		if err != nil {
			return nil, err
		}
		if !strings.Contains(qualitiesStr, JoinString(",", qualityStr, ",")) {
			return nil, errors.New(JoinString("channel: ", channel, ", reason: not support quality ", qualityStr))
		}
	} else {
		qualityStr, err = data.GetQuality(channel)
		if err != nil {
			return nil, err
		}
	}

	quality, err = strconv.Atoi(qualityStr)
	if err != nil {
		return nil, err
	}

	return &proc.QualityProcessor{quality, this.Cat}, nil
}

func (this *ProcChainBuilder) getWaterMarkProcessors(sourceType string, channel string, path string, params map[string]string) ([]proc.ImageProcessor, error) {
	processors := make([]proc.ImageProcessor, 2)
	//processors
	logoprocessor, err := this.getLogoWaterMarkProcessor(channel, params)
	if err != nil {
		return nil, err
	}
	if logoprocessor != nil {
		processors = append(processors, logoprocessor)
		log.Debug("add logo watermark processor")
	}
	nameprocessor, err := this.getNameWaterMarkProcessor(sourceType, channel, path, params)
	if err != nil {
		return nil, err
	}
	if nameprocessor != nil {
		processors = append(processors, nameprocessor)
		log.Debug("add name watermark processor")
	}

	return processors, nil
}
func (this *ProcChainBuilder) getLogoWaterMarkProcessor(channel string, params map[string]string) (proc.ImageProcessor, error) {
	dissolve := this.getLogoDissolve(channel, params)
	logodir, err := data.GetLogodir(channel)
	if err != nil {
		return nil, err
	}
	wn, _ := params["wn"]
	wl, _ := params["wl"]
	if wn == "" {
		defaultlogo, err := data.GetDefaultLogo(channel)
		if err != nil {
			return nil, err
		}
		arr := strings.Split(defaultlogo, ",")
		if len(arr) == 2 {
			wn = arr[0]
			wl = arr[1]
		}
	}
	if wn == "" {
		return nil, nil
	}
	//check watermarkname
	logonames, err := data.GetLogoNames(channel)
	if err != nil {
		return nil, err
	}
	if !strings.Contains(logonames, JoinString(",", wn, ",")) {
		return nil, errors.New(JoinString("channel: ", channel, ", reason: not support watermark ", wn))
	}
	//check size
	lesswidth, err := data.GetImagelesswidthForLogo(channel)
	if err != nil {
		return nil, err
	}
	lessheight, err := data.GetImagelessheightForLogo(channel)
	if err != nil {
		return nil, err
	}
	if lesswidth > 0 || lessheight > 0 {
		widthVal, _ := params[":3"]
		heightVal, _ := params[":4"]
		w, _ := strconv.ParseInt(widthVal, 10, 64)
		h, _ := strconv.ParseInt(heightVal, 10, 64)
		if !(w >= lesswidth && h >= lessheight) {
			return nil, nil
		}
	}
	l, err := strconv.Atoi(wl)
	if err != nil {
		l = 9
	}
	var path = logodir + wn + ".png"
	bts, err := GetImage("NFS", path, this.Cat)
	if err != nil {
		return nil, err
	}
	logo := &img4g.Image{Format: "png", Blob: bts, Cat: this.Cat}
	return &proc.WaterMarkProcessor{Logo: logo, Location: l, Dissolve: dissolve, Cat: this.Cat, WaterMarkType: "WaterMark"}, nil
}

func (this *ProcChainBuilder) getLogoDissolve(channel string, params map[string]string) int {
	if channel == Hotel || channel == Globalhotel {
		rotate, ok := params[":6"]
		if !ok {
			return 100
		}
		dissolves, _ := data.GetDissolves(channel)
		if !strings.Contains(dissolves, JoinString(",", rotate, ",")) {
			return 100
		}
		dissolve, _ := strconv.Atoi(rotate)
		return dissolve
	} else {
		return data.GetDissolve(channel)
	}
	return 100
}

func (this *ProcChainBuilder) getNameWaterMarkProcessor(sourceType string, channel string, path string, params map[string]string) (proc.ImageProcessor, error) {
	isMark, err := data.IsEnableNameLogo(channel)
	if err != nil {
		return nil, err
	}
	if isMark == false {
		return nil, nil
	}
	widthVal, _ := params[":3"]
	width, _ := strconv.ParseInt(widthVal, 10, 64)
	var logoname = this.getnamelogo(width)
	imagebts, err := GetImage(sourceType, path+logoname, this.Cat)
	if err != nil {
		return nil, nil
	}
	logo := &img4g.Image{Format: "png", Blob: imagebts, Cat: this.Cat}
	defer func() {
		logo.DestoryWand()
	}()
	if err := logo.CreateWand(); err != nil {
		return nil, err
	}
	logowidth, err := logo.GetWidth()
	if err != nil {
		return nil, err
	}
	l := 9
	if logowidth > width {
		l = 7
	}
	dissolve := this.getNameLogoDissolve(channel)
	return &proc.WaterMarkProcessor{Logo: logo, Location: l, Dissolve: dissolve, Cat: this.Cat, WaterMarkType: "NameWaterMark"}, nil
}

func (this *ProcChainBuilder) getNameLogoDissolve(channel string) int {
	return data.GetNamelogoDissolve(channel)
}

func (this *ProcChainBuilder) getnamelogo(width int64) string {
	if width <= 900 {
		return "_logo_14.png"
	}
	if width > 900 && width <= 1000 {
		return "_logo_16.png"
	}
	if width > 1000 && width <= 1100 {
		return "_logo_18.png"
	}
	return "_logo_20.png"
}
