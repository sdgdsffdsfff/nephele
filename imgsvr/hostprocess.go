package imgsvr

import (
	l4g "github.com/alecthomas/log4go"
	"github.com/ctripcorp/cat"
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
var hostPost string
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

func (this *HostProcessor) Run() {
	hostPost = strconv.Itoa(this.Port)
	threadcount := strconv.Itoa(this.ThreadCount)
	l4g.Debug(JoinString("Port:", hostPost, " threadcount:", threadcount, " nginxpath:", this.NginxPath, " nginxport:", this.NginxPort))
	defer func() {
		if err := recover(); err != nil {
			this.log(JoinString("hostprocessor->run(port:", hostPost, ",threadcount:", threadcount, ",nginxpath:", this.NginxPath, ",nginxport:", this.NginxPort, ",)"), err)
		}
		os.Exit(2)
	}()
	this.computePorts()
	if this.NginxPath != "" {
		if err := ModifyNginxconf(this.NginxPath, this.NginxPort, ports); err != nil {
			this.log(JoinString("hostprocessor->run->modifynginxcof(port:", hostPost, ",threadcount:", threadcount, ",nginxpath:", this.NginxPath, ",nginxport:", this.NginxPort, ",)"), err)
			return
		}
		if err := RestartNginx(this.NginxPath); err != nil {
			this.log(JoinString("hostprocessor->run->restartnginx(port:", hostPost, ",threadcount:", threadcount, ",nginxpath:", this.NginxPath, ",nginxport:", this.NginxPort, ",)"), err)
			return
		}
	}

	portstats = make(map[string]url.Values)
	for p, _ := range ports {
		this.runSubProc(p)
		l4g.Debug(JoinString("run sub proccess on port:", p))
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go this.monitorSubProcs()
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

func (this *HostProcessor) runSubProc(port string) {
	cmd := exec.Command("go", "run", "imgsvrd.go", "-s", port, hostPost)
	err := cmd.Start()
	if err != nil {
		this.log(JoinString("hostprocessor->runsubproc(port:", port, ",hostport:", hostPost, ",)"), err)
		return
	}
	return
}

func (this *HostProcessor) monitorSubProcs() {
	defer func() {
		if err := recover(); err != nil {
			l4g.Error(err)
			this.monitorSubProcs()
		}
	}()
	for {
		time.Sleep(2 * time.Second)
		l4g.Debug("monitor......")
		for port, countor := range ports {
			if port == "" {
				continue
			}

			if countor > 2 {
				l4g.Debug("restart port:" + port)
				err := KillProcessByPort(port)
				if err != nil {
					this.log(JoinString("hostprocessor->monitorsubprocs->killprocessbyport(port:", port, ")"), err)
				}
				this.runSubProc(port)
				ports[port] = 0
			} else {
				_, err := GetHttp(JoinString("http://127.0.0.1:", port, "/heartbeat/"))
				if err != nil {
					ports[port] = ports[port] + 1
					this.log(JoinString("hostprocessor->monitorsubprocs->get heartbeat(port:", port, ")"), err)
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
			l4g.Error(err)
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
		l4g.Debug("send cat heartbeat")
		stats1 := GetStats()
		data := url.Values{}
		data.Add("port", hostPost)
		for k, v := range stats1 {
			data.Add(k, v)
		}
		portstats[hostPost] = data

		tran := catinstance.NewTransaction("System", "Stats")
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
	l4g.Debug("listen and serve port:" + hostPost)
	//start server
	http.HandleFunc("/heartbeat/", this.heartbeatHandler)
	if err := http.ListenAndServe(":"+hostPost, nil); err != nil {
		this.log(JoinString("hostprocessor->listenHeartbeat(port:", hostPost, ")"), err)
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
	l4g.Debug(JoinString("get heartbeat from port[", port, "]"))

	w.Header().Set("Connection", "keep-alive")
	a := []byte(value)
	w.Header().Set("Content-Length", strconv.Itoa(len(a)))
	w.Write(a)
}

var cathost cat.Cat = cat.Instance()

func (this *HostProcessor) log(msg string, err interface{}) {
	l4g.Error("%s -- %s", msg, err)
	cathost.LogPanic(err)
}