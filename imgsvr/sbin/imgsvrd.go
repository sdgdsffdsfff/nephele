package main

import (
	l4g "github.com/alecthomas/log4go"
	"github.com/ctripcorp/nephele/imgsvr"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strconv"
)

func main() {
	runtime.GOMAXPROCS(1)
	l4g.LoadConfiguration("../conf/imgsvrd_log.xml")
	if len(os.Args) < 2 {
		l4g.Info("usage:params isn't invalid")
		os.Exit(1)
	}
	cmd := os.Args[1]
	if cmd == "-stop" {
		stop()
	}
	if cmd == "-nginx" {
		modifyNginx()
	}
	if cmd == "-h" {
		h()
	}
	if cmd == "-s" {
		s()
	}
	if cmd == "-reload" {
		reload()
	}
}

func modifyNginx() {
	if len(os.Args) < 3 {
		l4g.Info("usage:params isn't invalid")
		os.Exit(1)
	}
	nginxPath := os.Args[2]
	nginxPort := "80"
	if len(os.Args) > 3 {
		nginxPort = os.Args[3]
	}
	hostprocess := &imgsvr.HostProcessor{
		Port:        0,
		ThreadCount: 0,
		NginxPath:   nginxPath,
		NginxPort:   nginxPort,
	}
	hostprocess.ModifyNginxconf()
}

func reload() {
	if len(os.Args) < 3 {
		l4g.Info("usage:params isn't invalid")
		os.Exit(1)
	}
	portstr := os.Args[2]
	port, err := strconv.Atoi(portstr)
	if err != nil {
		l4g.Error(err)
		os.Exit(1)
	}
	hostprocess := &imgsvr.HostProcessor{
		Port:        port,
		ThreadCount: 0,
		NginxPath:   "",
		NginxPort:   "",
	}
	hostprocess.ReloadConf()
}

func h() {
	if len(os.Args) < 3 {
		l4g.Info("usage:params isn't invalid")
		os.Exit(1)
	}
	portstr := os.Args[2]
	port, err := strconv.Atoi(portstr)
	if err != nil {
		l4g.Error(err)
		os.Exit(1)
	}
	threadcount, nginxpath, nginxport := getArgs()
	hostprocess := &imgsvr.HostProcessor{
		Port:        port,
		ThreadCount: threadcount,
		NginxPath:   nginxpath,
		NginxPort:   nginxport,
	}
	hostprocess.Run()
}
func s() {
	if len(os.Args) < 3 {
		l4g.Info("usage:params isn't invalid")
		os.Exit(1)
	}
	portstr := os.Args[2]
	var hostport string = ""
	if len(os.Args) > 3 {
		_, err := strconv.Atoi(os.Args[3])
		if err == nil {
			hostport = os.Args[3]
		}
	}
	subprocess := &imgsvr.SubProcessor{
		Port:     portstr,
		HostPort: hostport,
	}
	subprocess.Run()
}
func stop() {
	hostprocess := &imgsvr.HostProcessor{
		Port:        0,
		ThreadCount: 0,
		NginxPath:   "",
		NginxPort:   "",
	}
	hostprocess.Stop()
}

func getArgs() (int, string, string) {
	var threadCount int = 0
	var path string = ""
	var nginxPort string
	if len(os.Args) > 3 {
		path = os.Args[3]
		if len(os.Args) > 4 {
			nginxPort = os.Args[4]
			if len(os.Args) > 5 {
				c, _ := strconv.Atoi(os.Args[5])
				if c > 0 {
					threadCount = c
				}
			}
		}
	}
	return threadCount, path, nginxPort
}
