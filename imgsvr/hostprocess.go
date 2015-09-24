package imgsvr

import (
	"fmt"
	log "github.com/ctripcorp/nephele/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	cat "github.com/ctripcorp/nephele/Godeps/_workspace/src/github.com/ctripcorp/cat.go"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"time"
)

type HostProcessor struct {
	Port        int
	ThreadCount int
	NginxPath   string
	NginxPort   string
}

var ports map[string]int
var hostPort string
var portstats map[string]url.Values

func (this *HostProcessor) Stop() {
	cmd := exec.Command("pkill", "imgsvrd")
	cmd.Output()
}

func (this *HostProcessor) ReloadConf() {
	//data.Reload()
	this.computePorts()
	for p, _ := range ports {
		if p != "" {
			GetHttp(JoinString("http://127.0.0.1:", p, "/reload/"))
		}
	}
}

func (this *HostProcessor) ModifyNginxconf() {
	this.computePorts()
	if this.NginxPath != "" {
		if err := ModifyNginxconf(this.NginxPath, this.NginxPort, ports); err != nil {
			log.WithFields(log.Fields{
				"nginxPath": this.NginxPath,
				"nginxPort": this.NginxPort,
				"type":      "Daemon.ModifyNginxError",
			}).Error(err.Error())
			LogErrorEvent(CatInstance, "Daemon.ModifyNginxError", err.Error())
			return
		}
		if err := RestartNginx(this.NginxPath); err != nil {
			log.WithFields(log.Fields{
				"nginxPath": this.NginxPath,
				"nginxPort": this.NginxPort,
				"type":      "Daemon.RestartNginxError",
			}).Error(err.Error())
			LogErrorEvent(CatInstance, "Daemon.RestartNginxError", err.Error())
			return
		}
	}
}

func (this *HostProcessor) Run() {
	hostPort = strconv.Itoa(this.Port)
	threadcount := strconv.Itoa(this.ThreadCount)
	log.WithFields(log.Fields{
		"hostPort":    hostPort,
		"threadCount": threadcount,
		"nginxPath":   this.NginxPath,
		"nginxPort":   this.NginxPort,
	}).Debug("run host process")

	defer func() {
		if p := recover(); p != nil {
			log.WithFields(log.Fields{
				"hostPort":    hostPort,
				"threadCount": threadcount,
				"type":        "DaemonProcess.RunPanic",
			}).Error(fmt.Sprintf("%v", p))
			LogErrorEvent(CatInstance, "DaemonProcess.RunPanic", fmt.Sprintf("%v", p))
		}
		os.Exit(2)
	}()
	func() {
		Cat := cat.Instance()
		tran := Cat.NewTransaction("System", Reboot)
		defer func() {
			tran.SetStatus("0")
			tran.Complete()
		}()
		LogEvent(Cat, Reboot, JoinString(GetIP(), ":", hostPort), nil)
	}()

	this.computePorts()
	if this.NginxPath != "" {
		if err := ModifyNginxconf(this.NginxPath, this.NginxPort, ports); err != nil {
			log.WithFields(log.Fields{
				"nginxPath": this.NginxPath,
				"nginxPort": this.NginxPort,
				"type":      "DaemonProcess.ModifyNginxError",
			}).Error(err.Error())
			LogErrorEvent(CatInstance, "DaemonProcess.ModifyNginxError", err.Error())
			return
		}
		if err := RestartNginx(this.NginxPath); err != nil {
			log.WithFields(log.Fields{
				"nginxPath": this.NginxPath,
				"type":      "DaemonProcess.RestartNginxError",
			}).Error(err.Error())
			LogErrorEvent(CatInstance, "DaemonProcess.RestartNginxError", err.Error())
			return
		}
	}

	portstats = make(map[string]url.Values)
	for p, _ := range ports {
		this.startWorkerProcess(p)
		log.WithFields(log.Fields{
			"port": p,
		}).Debug("run subproccess on port")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go this.monitorWorkerProcesses()
	go this.sendCatHeartBeat()
	go this.listenHeartbeat()
	<-c //Interupt signal coming
	os.Exit(0)
}

func (this *HostProcessor) computePorts() {
	if this.ThreadCount < 1 || this.ThreadCount > runtime.NumCPU() {
		this.ThreadCount = runtime.NumCPU()
	}
	ports = make(map[string]int, this.ThreadCount)
	for i := 0; i < this.ThreadCount; i++ {
		ports[strconv.Itoa(this.Port+i+1)] = 0
	}
}

func (this *HostProcessor) startWorkerProcess(port string) {
	cmd := exec.Command("go", "run", "imgsvrd.go", "-s", port, hostPort)
	err := cmd.Start()
	if err != nil {
		log.WithFields(log.Fields{
			"port":     port,
			"hostPort": hostPort,
			"type":     "DaemonProcess.StartWorkerError",
		}).Error(err.Error())
		LogErrorEvent(CatInstance, "DaemonProcess.StartWorkerError", err.Error())
		return
	}
	return
}

func (this *HostProcessor) monitorWorkerProcesses() {
	defer func() {
		if p := recover(); p != nil {
			log.WithFields(log.Fields{
				"type": "DaemonProcess.MonitorPanic",
			}).Error(fmt.Sprintf("%v", p))
			LogErrorEvent(CatInstance, "DaemonProcess.MonitorPanic", fmt.Sprintf("%v", p))
			this.monitorWorkerProcesses()
		}
	}()
	time.Sleep(8 * time.Second) //sleep ,wait sub process run
	for {
		time.Sleep(2 * time.Second)
		log.Debug("monitor......")
		for port, countor := range ports {
			if port == "" {
				continue
			}

			if countor > 2 {
				log.WithFields(log.Fields{
					"port": port,
				}).Debug("restart port")
				err := KillProcessByPort(port)
				if err != nil {
					log.WithFields(log.Fields{
						"port": port,
						"type": "DaemonProcess.KillProcessError",
					}).Error(err.Error())
					LogErrorEvent(CatInstance, "DaemonProcess.KillProcessError", err.Error())
				}
				this.startWorkerProcess(port)
				ports[port] = 0
			} else {
				_, err := GetHttp(JoinString("http://127.0.0.1:", port, "/heartbeat/"))
				if err != nil {
					ports[port] = ports[port] + 1
					log.WithFields(log.Fields{
						"port": port,
						"type": "WorkerProcess.HeartbeatError",
					}).Error(err.Error())
					LogErrorEvent(CatInstance, "WorkerProcess.HeartbeatError", err.Error())
				} else {
					ports[port] = 0
				}
			}
		}
	}
}

func (this *HostProcessor) sendCatHeartBeat() {
	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("%v", err))
			this.sendCatHeartBeat()
		}
	}()
	ip := GetIP()
	second := time.Now().Second()
	if second < 29 {
		sleep := time.Duration((29 - second) * 1000000000)
		time.Sleep(sleep)
	}

	catinstance := cat.Instance()
	for {
		log.Debug("send cat heartbeat")
		stats1 := GetStatus()
		data := url.Values{}
		data.Add("port", hostPort)
		for k, v := range stats1 {
			data.Add(k, v)
		}
		portstats[hostPort] = data

		tran := catinstance.NewTransaction("System", "Status")
		h := catinstance.NewHeartbeat("HeartBeat", ip)
		for _, heart := range portstats {
			if heart == nil {
				continue
			}
			port := heart.Get("port")
			if port == "" {
				continue
			}
			h.Set("System", JoinString("Alloc_", port), heart.Get("Alloc"))
			h.Set("System", JoinString("TotalAlloc_", port), heart.Get("TotalAlloc"))
			h.Set("System", JoinString("Sys_", port), heart.Get("Sys"))
			h.Set("System", JoinString("Mallocs_", port), heart.Get("Mallocs"))
			h.Set("System", JoinString("Frees_", port), heart.Get("Frees"))
			h.Set("System", JoinString("OtherSys_", port), heart.Get("OtherSys"))
			h.Set("System", JoinString("PauseNs_", port), heart.Get("PauseNs"))

			h.Set("HeapUsage", JoinString("HeapAlloc_", port), heart.Get("HeapAlloc"))
			h.Set("HeapUsage", JoinString("HeapSys_", port), heart.Get("HeapSys"))
			h.Set("HeapUsage", JoinString("HeapIdle_", port), heart.Get("HeapIdle"))
			h.Set("HeapUsage", JoinString("HeapInuse_", port), heart.Get("HeapInuse"))
			h.Set("HeapUsage", JoinString("HeapReleased_", port), heart.Get("HeapReleased"))
			h.Set("HeapUsage", JoinString("HeapObjects_", port), heart.Get("HeapObjects"))

			h.Set("GC", JoinString("NextGC_", port), heart.Get("NextGC"))
			h.Set("GC", JoinString("LastGC_", port), heart.Get("LastGC"))
			h.Set("GC", JoinString("NumGC_", port), heart.Get("NumGC"))
			portstats[port] = nil
		}
		h.SetStatus("0")
		h.Complete()
		tran.SetStatus("0")
		tran.Complete()
		second = time.Now().Second()
		sleep := time.Duration((90 - second) * 1000000000)
		time.Sleep(sleep)
	}
}

func (this *HostProcessor) listenHeartbeat() {
	log.WithFields(log.Fields{
		"hostPort": hostPort,
	}).Debug("listen and serve port")
	//start server
	http.HandleFunc("/heartbeat/", this.heartbeatHandler)
	if err := http.ListenAndServe(":"+hostPort, nil); err != nil {
		log.WithFields(log.Fields{
			"hostPort": hostPort,
			"type":     "DaemonProcess.ListenAndServeError",
		}).Error(err.Error())
		LogErrorEvent(CatInstance, "DaemonProcess.ListenAndServeError", err.Error())
		os.Exit(1)
	}
}

func (this *HostProcessor) heartbeatHandler(w http.ResponseWriter, request *http.Request) {
	port := request.FormValue("port")
	var value string = "0"
	if port != "" {
		portstats[port] = request.Form
		value = "1"
	}
	log.WithFields(log.Fields{
		"port": port,
	}).Debug("get heartbeat from port")

	w.Header().Set("Connection", "keep-alive")
	a := []byte(value)
	w.Header().Set("Content-Length", strconv.Itoa(len(a)))
	w.Write(a)
}
