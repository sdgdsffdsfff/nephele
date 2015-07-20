package imgsvr

import (
	l4g "github.com/alecthomas/log4go"
	"github.com/ctripcorp/cat"
	"github.com/ctripcorp/nephele/imgsvr/data"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"time"
)

type SubProcessor struct {
	Port     string
	HostPort string
}

func (this *SubProcessor) Run() {
	defer func() {
		if err := recover(); err != nil {
			this.log(JoinString("sbuprocessor->run(port:", this.Port, ",hostport:", this.HostPort, ")"), err)
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	go CycleHandleImage()
	go this.listenHttp()
	if this.HostPort != "" {
		go this.sendStats()
	}
	<-c
	os.Exit(0)
}
func (this *SubProcessor) listenHttp() {
	handler := &Handler{}
	http.Handle("/images/", handler)
	http.HandleFunc("/heartbeat/", this.handleHeartbeart)
	http.HandleFunc("/reload/", this.reload)
	l4g.Debug("starthttp port:" + this.Port)
	err := http.ListenAndServe(":"+this.Port, nil)
	if err != nil {
		panic(err)
	}
}

func (this *SubProcessor) reload(w http.ResponseWriter, request *http.Request) {
	err := data.Reload()
	var value string = "1"
	w.Header().Set("Connection", "keep-alive")
	if err != nil {
		value = "0"
	}
	a := []byte(value)
	w.Header().Set("Content-Length", strconv.Itoa(len(a)))
	w.Write(a)
}

func (this *SubProcessor) handleHeartbeart(w http.ResponseWriter, request *http.Request) {
	var value string = "1"
	w.Header().Set("Connection", "keep-alive")
	a := []byte(value)
	w.Header().Set("Content-Length", strconv.Itoa(len(a)))
	w.Write(a)
}
func (this *SubProcessor) sendStats() {
	defer func() {
		if err := recover(); err != nil {
			this.log(JoinString("subprocessor->sendStats(port:", this.Port, ")"), err)
			this.sendStats()
		}
	}()
	for {
		time.Sleep(10 * time.Second)
		uri := JoinString("http://localhost:", this.HostPort, "/heartbeat/")
		l4g.Debug(JoinString("port:", this.Port, " | ", uri))
		status := GetStats()
		data := url.Values{}
		data.Add("port", this.Port)
		for k, v := range status {
			data.Add(k, v)
		}
		_, err := PostHttp(uri, data)
		if err != nil {
			// handle error
			this.log(JoinString("subprocessor->sendStats->posthttp(port:", this.Port, ")"), err)
		}
	}
}

var catsub cat.Cat = cat.Instance()

func (this *SubProcessor) log(msg string, err interface{}) {
	l4g.Error("%s -- %s", msg, err)
	catsub.LogPanic(err)
}