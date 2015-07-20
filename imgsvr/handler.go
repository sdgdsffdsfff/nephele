package imgsvr

import (
	"errors"
	l4g "github.com/alecthomas/log4go"
	"github.com/ctripcorp/cat"
	"github.com/ctripcorp/nephele/imgsvr/img4g"
	"github.com/ctripcorp/nephele/imgsvr/proc"
	"github.com/ctripcorp/nephele/imgsvr/storage"
	"github.com/ctripcorp/nephele/util"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var legalUrl = util.RegexpExt{regexp.MustCompile("^/images/(.*?)_(R|C|Z|W)_([0-9]+)_([0-9]+)(_R([0-9]+))?(_C([a-zA-Z]+))?(_Q(?P<n0>[0-9]+))?(_M((?P<wn>[a-zA-Z0-9]+)(_(?P<wl>[1-9]))?))?.(?P<ext>jpg|jpeg|gif|png|Jpg)$")}
var forbiddenUrl = util.RegexpExt{regexp.MustCompile("^/images/fd/([a-zA-Z]+)/([a-zA-Z0-9]+)/(.*?)_Source.(?P<ext>jpg|jpeg|gif|png|Jpg)$")}
var proxyPassUrl = util.RegexpExt{regexp.MustCompile("^/images/fd/([a-zA-Z]+)/([a-zA-Z0-9]+)/(.*?).(?P<ext>jpg|jpeg|gif|png|Jpg)$")}

type Handler struct {
	ChainBuilder *ProcChainBuilder
}

func (handler *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	Cat := cat.Instance()
	handler.ChainBuilder = &ProcChainBuilder{Cat}
	tran := Cat.NewTransaction("Image.Request", "Request")
	var (
		ok  bool
		err error
	)
	defer func() {
		p := recover()
		if p != nil {
			l4g.Error(p)
			Cat.LogPanic(p)
		}
		if !ok {
			tran.SetStatus("1")
		}
		if err != nil {
			l4g.Error(err)
			Cat.LogError(err)
			tran.SetStatus(err)
		}
		if p != nil {
			tran.SetStatus(p)
		} else {
			tran.SetStatus("0")
		}
		tran.Complete()

		if !ok || err != nil || p != nil {
			http.Error(writer, http.StatusText(404), 404)
		}
	}()
	uri := request.URL.String()
	tran.AddData("url", uri)
	params, ok1 := legalUrl.FindStringSubmatchMap(uri)
	ok = ok1
	if !ok1 {
		l4g.Error("url.unexpected.mark(url:%s)", uri)
		Cat.LogPanic(JoinString("url.unexpected.mark url:", uri))
		return
	}
	store, err1 := FindStorage(params)
	err = err1
	if err1 != nil {
		l4g.Error("%s; rcv(url:%s)", err, request.URL.String())
		return
	}
	chain, err1 := handler.ChainBuilder.Build(params)
	err = err1
	if err1 != nil {
		return
	}

	bts, err1 := store.GetImage()
	err = err1
	if err1 != nil {
		return
	}
	l4g.Debug("get image length(%d) rcv(url:%s)", len(bts), request.URL.String())
	format, _ := params["ext"]
	img := &img4g.Image{Blob: bts, Format: format}
	status := make(chan bool)
	imgHd := imgHandle{img, chain, status}
	imgHandleChan <- imgHd
	timeout := make(chan bool, 1)
	ok = true
	go func() {
		time.Sleep(10 * time.Second) //wait 10s
		timeout <- true
	}()
	select {
	case ok = <-status:
		if !ok {
			return
		}
	case <-timeout:
		ok = false
		return
	}
	writer.Header().Set("Content-Type", "image/"+format)
	writer.Header().Set("Content-Length", strconv.Itoa(len(img.Blob)))
	if _, err = writer.Write(img.Blob); err != nil {
		l4g.Error(err)
		return
	}
}

func CycleHandleImage() {
	defer func() {
		if err := recover(); err != nil {
			log("imagehandler->cyclehandleimage", err)
			go CycleHandleImage()
		}
	}()

	for {
		status := true
		imgHd := <-imgHandleChan
		chain := imgHd.chain
		image := imgHd.inImg
		if err := chainProcImg(chain, image); err != nil {

			log("imagehandler->cyclehandleimage->for", err)
			status = false
		}
		timeout := make(chan bool, 1)
		go func() {
			time.Sleep(1 * 1e9) //wait 1s
			timeout <- true
		}()
		select {
		case imgHd.status <- status:
		case <-timeout:
		}
	}
}

func chainProcImg(chain *proc.ProcessorChain, img *img4g.Image) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
			log("imagehandler->chainprocimg", r)
		}
	}()
	defer img.DestoryWand()
	if err = img.CreateWand(); err != nil {
		return err
	}
	if err = chain.Process(img); err != nil {
		return err
	}
	if err = img.WriteImageBlob(); err != nil {
		return err
	}
	return nil
}

func FindStorage(params map[string]string) (storage.Storage, error) {
	srcPath, ok := params[":1"]
	if !ok {
		return nil, errors.New("Url.UnExpected")
	}
	format, ok := params["ext"]
	if !ok {
		return nil, errors.New("Image.Ext.Invalid()")
	}
	sourceType, _, path := ParseUri(srcPath)
	return GetStorage(sourceType, path+"."+format)
}

var catimage cat.Cat = cat.Instance()

func log(msg string, err interface{}) {
	l4g.Error("%s -- %s", msg, err)
	catimage.LogPanic(err)
}
