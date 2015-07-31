package imgsvr

import (
	"errors"
	"fmt"
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

var (
	legalUrl     = util.RegexpExt{regexp.MustCompile("^/images/(.*?)_(R|C|Z|W)_([0-9]+)_([0-9]+)(_R([0-9]+))?(_C([a-zA-Z]+))?(_Q(?P<n0>[0-9]+))?(_M((?P<wn>[a-zA-Z0-9]+)(_(?P<wl>[1-9]))?))?.(?P<ext>jpg|jpeg|gif|png|Jpg)$")}
	forbiddenUrl = util.RegexpExt{regexp.MustCompile("^/images/fd/([a-zA-Z]+)/([a-zA-Z0-9]+)/(.*?)_Source.(?P<ext>jpg|jpeg|gif|png|Jpg)$")}
	proxyPassUrl = util.RegexpExt{regexp.MustCompile("^/images/fd/([a-zA-Z]+)/([a-zA-Z0-9]+)/(.*?).(?P<ext>jpg|jpeg|gif|png|Jpg)$")}
)

type Handler struct {
	ChainBuilder *ProcChainBuilder
}

func (handler *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	Cat := cat.Instance()
	handler.ChainBuilder = &ProcChainBuilder{Cat}
	tran := Cat.NewTransaction("Image.Request", "Request")
	uri := request.URL.String()
	var (
		err error
	)
	defer func() {
		p := recover()
		if p != nil {
			l4g.Error("%s; rcv(url:%s)", p, uri)
			Cat.LogPanic(p)
			tran.SetStatus(p)
		}
		if err != nil {
			tran.SetStatus(err)
		}
		if p == nil && err == nil {
			tran.SetStatus("0")
			tran.Complete()
		} else {
			tran.Complete()
			http.Error(writer, http.StatusText(404), 404)
		}
	}()
	data := map[string]string{
		"referer":    request.Referer(),
		"proto":      request.Proto,
		"agent":      request.UserAgent(),
		"remoteaddr": request.RemoteAddr,
	}
	LogEvent(Cat, "URL", "URL.client", data)

	data1 := map[string]string{
		"Http": request.Method + " " + uri,
	}
	LogEvent(Cat, "URL", "URL.method", data1)
	params, ok1 := legalUrl.FindStringSubmatchMap(uri)
	if !ok1 {
		err = errors.New("uri.parseerror")
		LogEvent(Cat, "Error", "uri.parseerror", make(map[string]string))
		l4g.Error("%s; rcv(url:%s)", err, uri)
		return
	}
	store, storagetype, err1 := FindStorage(params)
	if err1 != nil {
		err = errors.New("storage.parseerror")
		LogEvent(Cat, "Error", "storage.parseerror", map[string]string{"detail": err1.Error()})
		l4g.Error("%s; rcv(url:%s)", err1, uri)
		return
	}
	chain, err1 := handler.ChainBuilder.Build(params)
	if err1 != nil {
		LogEvent(Cat, "Error", "procchain.build", map[string]string{"detail": err1.Error()})
		err = errors.New("procchain.builderror")
		l4g.Error("%s; rcv(url:%s)", err1, uri)
		return
	}

	getimagetran := Cat.NewTransaction("Image.Storage", storagetype)
	bts, err1 := store.GetImage()
	if err1 != nil {
		getimagetran.AddData("getimage", err1.Error())
		LogErrorEvent(Cat, storagetype+".readerror", err1.Error())
		err = errors.New(storagetype + ".readerror")
		l4g.Error("%s -- %s; rcv(url:%s)", "getimage", err1, uri)
		getimagetran.SetStatus(err)
		return
	} else {
		getimagetran.SetStatus("0")
	}
	getimagetran.Complete()
	size := len(bts)
	sizestr := strconv.Itoa(size)
	tran.AddData("size", sizestr)
	Cat.LogEvent("Image.Size", GetImageSizeDistribution(size))

	l4g.Debug("get image length(%d) rcv(url:%s)", size, request.URL.String())
	format, _ := params["ext"]
	img := &img4g.Image{Blob: bts, Format: format}

	status := make(chan bool, 1)
	imgHd := imgHandle{img, chain, status, Cat}
	imgHandleChan <- imgHd

	select {
	case ok := <-status:
		if !ok {
			err = errors.New("image.processerror")
			LogErrorEvent(Cat, "image.processerror", "")
			l4g.Error("%s; rcv(url:%s)", err, uri)
			return
		}
	case <-time.After(time.Second * 5):
		err = errors.New("image.processtimeout")
		LogErrorEvent(Cat, "image.processtimeout", "")
		l4g.Error("%s; rcv(url:%s)", err, uri)
		return
	}
	writer.Header().Set("Content-Type", "image/"+format)
	writer.Header().Set("Content-Length", strconv.Itoa(len(img.Blob)))
	l4g.Debug("final size->>>" + strconv.Itoa(len(img.Blob)))
	if _, err1 = writer.Write(img.Blob); err1 != nil {
		l4g.Error(err1)
		tran.AddData("response", err1.Error())
		err = errors.New("response.writeerror")
		LogErrorEvent(Cat, "response.writeerror", err1.Error())
		return
	}
}

func CycleHandleImage() {
	defer func() {
		if err := recover(); err != nil {
			l4g.Error("%s -- %s", "handleimage.recovererror", err)
			LogErrorEvent(CatInstance, "handleimage.recovererror", fmt.Sprintf("%v", err))
			go CycleHandleImage()
		}
	}()

	for {
		status := true
		imgHd := <-imgHandleChan
		chain := imgHd.chain
		image := imgHd.inImg
		if err := chainProcImg(imgHd.CatInstance, chain, image); err != nil {
			l4g.Error("%s -- %s", "handleimage.processerror", err)
			LogErrorEvent(imgHd.CatInstance, "handleimage.processerror", err.Error())
			status = false
		}
		imgHd.status <- status
	}
}

func chainProcImg(catinstance cat.Cat, chain *proc.ProcessorChain, img *img4g.Image) (err error) {
	defer func() {
		if r := recover(); r != nil {
			l4g.Error("%s -- %s", "processimage.recovererror", err)
			LogErrorEvent(catinstance, "processimage.recovererror", fmt.Sprintf("%v", r))
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

func FindStorage(params map[string]string) (storage.Storage, string, error) {
	srcPath, ok := params[":1"]
	if !ok {
		return nil, "", errors.New("Url.UnExpected")
	}
	format, ok := params["ext"]
	if !ok {
		return nil, "", errors.New("Image.Ext.Invalid()")
	}
	sourceType, _, path := ParseUri(srcPath)
	s, err := GetStorage(sourceType, path+"."+format)
	return s, sourceType, err
}
