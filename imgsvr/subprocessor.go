package imgsvr

import (
	"fmt"
	log "github.com/ctripcorp/nephele/Godeps/_workspace/src/github.com/Sirupsen/logrus"
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
		if p := recover(); p != nil {
			log.WithFields(log.Fields{
				"hostPort":   this.HostPort,
				"workerPort": this.Port,
				"type":       "Worker.RunPanic",
			}).Error(fmt.Sprintf("%v", p))
			LogErrorEvent(CatInstance, "Worker.RunPanic", fmt.Sprintf("%v", p))
		}
	}()
	WorkerPort = this.Port
	LogEvent(CatInstance, Reboot, JoinString(GetIP(), ":", this.Port), nil)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	go CycleHandleImage()
	go this.listenHttp()
	if this.HostPort != "" {
		go this.sendStatus()
	}
	<-c
	os.Exit(0)
}
func (this *SubProcessor) listenHttp() {
	handler := &Handler{}
	http.Handle("/images/", handler)
	http.HandleFunc("/heartbeat/", this.handleHeartbeart)
	http.HandleFunc("/reload/", this.reload)
	log.WithFields(log.Fields{
		"port": this.Port,
	}).Debug("http start port")
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
func (this *SubProcessor) sendStatus() {
	defer func() {
		if p := recover(); p != nil {
			log.WithFields(log.Fields{
				"port": this.Port,
				"type": "Worker.SendStatusPanic",
			}).Error(fmt.Sprintf("%v", p))
			LogErrorEvent(CatInstance, "Worker.SendStatusPanic", fmt.Sprintf("%v", p))
			this.sendStatus()

		}
	}()
	for {
		time.Sleep(10 * time.Second)
		uri := JoinString("http://localhost:", this.HostPort, "/heartbeat/")
		log.WithFields(log.Fields{
			"port": this.Port,
			"uri":  uri,
		}).Debug("begin send status")
		status := GetStatus()
		data := url.Values{}
		data.Add("port", this.Port)
		for k, v := range status {
			data.Add(k, v)
		}
		_, err := PostHttp(uri, data)
		if err != nil {
			// handle error
			log.WithFields(log.Fields{
				"port": this.Port,
				"type": "Worker.SendStatusError",
			}).Error(err.Error())
			LogErrorEvent(CatInstance, "Worker.SendStatusError", err.Error())
		}
	}
}
