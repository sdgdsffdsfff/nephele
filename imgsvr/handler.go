package imgsvr

import (
	"errors"
	"fmt"
	log "github.com/ctripcorp/nephele/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	cat "github.com/ctripcorp/nephele/Godeps/_workspace/src/github.com/ctripcorp/cat.go"
	"github.com/ctripcorp/nephele/imgsvr/img4g"
	"github.com/ctripcorp/nephele/imgsvr/proc"
	"github.com/ctripcorp/nephele/imgsvr/storage"
	"github.com/ctripcorp/nephele/util"
	"net/http"
	"regexp"
	"strconv"
	"strings"
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
	uri := request.URL.String()
	tran := Cat.NewTransaction("URL", getShortUri(uri))
	var (
		err       error
		isSuccess bool = true
	)
	defer func() {
		p := recover()
		if p != nil {
			logErrWithUri(uri, fmt.Sprintf("%v", p), "errorLevel")
			Cat.LogPanic(p)
			tran.SetStatus(p)
		}

		if isSuccess {
			tran.SetStatus("0")
			tran.Complete()
		} else {
			tran.SetStatus(err)
			tran.Complete()
		}
		if p != nil || err != nil {
			http.Error(writer, http.StatusText(404), 404)
		}
	}()

	LogEvent(Cat, "URL", "URL.Client", map[string]string{
		"clientip": GetClientIP(request),
		"serverip": GetIP(),
		"proto":    request.Proto,
		"referer":  request.Referer(),
		//"agent":    request.UserAgent(),
	})

	LogEvent(Cat, "URL", "URL.Method", map[string]string{
		"Http": request.Method + " " + uri,
	})

	LogEvent(Cat, "UpstreamProcess", JoinString(GetIP(), ":", WorkerPort), nil)

	params, ok1 := legalUrl.FindStringSubmatchMap(uri)
	if !ok1 {
		err = errors.New("URI.ParseError")
		logErrWithUri(uri, err.Error(), "warnLevel")
		LogErrorEvent(Cat, "URI.ParseError", "")
		return
	}
	//parse storage from url parameters
	store, storagetype, err1 := FindStorage(params, Cat)
	if err1 != nil {
		err = errors.New("Storage.ParseError")
		logErrWithUri(uri, err1.Error(), "warnLevel")
		LogErrorEvent(Cat, "Storage.ParseError", err1.Error())
		return
	}
	//parse handlers chain from url parameters
	chain, buildErr := handler.ChainBuilder.Build(params)
	if buildErr != nil {
		err = errors.New(buildErr.Type())
		logErrWithUri(uri, buildErr.Error(), "warnLevel")
		LogErrorEvent(Cat, buildErr.Type(), buildErr.Error())
		return
	}
	//download image from storage
	var bts []byte
	func() {
		type storageError interface {
			Error() string
			Normal() bool //is normal error?
			Type() string //error type
		}

		var err1 error
		getimagetran := Cat.NewTransaction("Storage", storagetype)
		defer func() {
			if err1 != nil {
				logErrWithUri(uri, err1.Error(), "errorLevel")
				e, ok := err1.(storageError)
				if ok && e.Normal() {
					err = errors.New(fmt.Sprintf("%v.%v", storagetype, e.Type()))
					LogErrorEvent(Cat, fmt.Sprintf("%v.%v", storagetype, e.Type()), e.Error())
				} else {
					err = errors.New(storagetype + ".UnExpectedError")
					LogErrorEvent(Cat, err.Error(), err1.Error())
					isSuccess = false
				}
			} else if len(bts) == 0 {
				err = errors.New(storagetype + ".ImgLenZero")
				LogErrorEvent(Cat, err.Error(), "recv image length is 0")
				logErrWithUri(uri, "recv image length is 0", "warnLevel")
			}
			if isSuccess {
				getimagetran.SetStatus("0")
			} else {
				getimagetran.SetStatus(err)
			}
			getimagetran.Complete()
		}()
		bts, err1 = store.GetImage()
	}()
	if err != nil {
		return
	}
	size := len(bts)
	sizestr := strconv.Itoa(size)
	tran.AddData("size", sizestr)
	Cat.LogEvent("Size", GetImageSizeDistribution(size))

	log.WithFields(log.Fields{
		"size": size,
		"uri":  uri,
	}).Debug("recv image length")
	format, _ := params["ext"]
	img := &img4g.Image{Blob: bts, Format: format, Cat: Cat}

	rspChan := make(chan bool, 1)
	task := &nepheleTask{inImg: img, chain: chain, rspChan: rspChan, CatInstance: Cat, canceled: false}
	taskChan <- task

	select {
	case ok := <-rspChan:
		if !ok {
			err = errors.New("ProcessError")
			isSuccess = false
			logErrWithUri(uri, err.Error(), "errorLevel")
			return
		}
	case <-time.After(time.Second * 5):
		task.SetCanceled()
		err = errors.New("ProcessTimeout")
		logErrWithUri(uri, err.Error(), "errorLevel")
		isSuccess = false
		LogErrorEvent(Cat, "ProcessTimeout", "")
		return
	}

	writer.Header().Set("Content-Type", "image/"+format)
	writer.Header().Set("Content-Length", strconv.Itoa(len(img.Blob)))
	writer.Header().Set("Last-Modified", "2015/1/1 01:01:01")
	log.WithFields(log.Fields{
		"size": size,
		"uri":  uri,
	}).Debug("final image size")
	if _, err1 = writer.Write(img.Blob); err1 != nil {
		logErrWithUri(uri, err1.Error(), "errorLevel")
		err = errors.New("Response.WriteError")
		LogErrorEvent(Cat, "Response.Writeerror", err1.Error())
		isSuccess = false
	}
}

func CycleHandleImage() {
	defer func() {
		if r := recover(); r != nil {
			log.WithFields(log.Fields{
				"type": "HandleImagePanic",
			}).Error(fmt.Sprintf("%v", r))
			LogErrorEvent(CatInstance, "HandleImagePanic", fmt.Sprintf("%v", r))
			go CycleHandleImage()
		}
	}()

	for {
		status := true
		//get a task from task chan
		task := <-taskChan
		if task.GetCanceled() {
			continue
		}
		chain := task.chain
		image := task.inImg
		if err := chainProcImg(task.CatInstance, chain, image); err != nil {
			log.WithFields(log.Fields{
				"type": "ProcessError",
			}).Error(err.Error())
			LogErrorEvent(task.CatInstance, "ProcessError", err.Error())
			status = false
		}
		task.rspChan <- status
	}
}

func chainProcImg(catinstance cat.Cat, chain *proc.ProcessorChain, img *img4g.Image) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.WithFields(log.Fields{
				"type": "ProcessImage.Panic",
			}).Error(fmt.Sprintf("%v", r))
			LogErrorEvent(catinstance, "ProcessImage.Panic", fmt.Sprintf("%v", r))
		}
	}()
	defer img.DestoryWand()
	if err = img.CreateWand(); err != nil {
		return
	}
	if err = chain.Process(img); err != nil {
		return
	}
	err = img.WriteImageBlob()
	return
}

func FindStorage(params map[string]string, Cat cat.Cat) (storage.Storage, string, error) {
	srcPath, ok := params[":1"]
	if !ok {
		return nil, "", errors.New("Url.UnExpected")
	}
	format, ok := params["ext"]
	if !ok {
		return nil, "", errors.New("Image.Ext.Invalid()")
	}
	sourceType, _, path := ParseUri(srcPath)
	s, err := GetStorage(sourceType, path+"."+format, Cat)
	return s, sourceType, err
}

func getShortUri(uri string) string {
	arr := strings.Split(uri, "/")
	if len(arr) < 4 {
		return uri
	}
	if arr[2] == "fd" || arr[2] == "t1" {
		return JoinString("/", arr[1], "/", arr[2], "/", arr[3])
	} else {
		return JoinString("/", arr[1], "/", arr[2])
	}
}
