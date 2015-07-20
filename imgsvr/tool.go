package imgsvr

import (
	"bytes"
	"errors"
	"github.com/ctripcorp/nephele/imgsvr/data"
	"github.com/ctripcorp/nephele/imgsvr/img4g"
	"github.com/ctripcorp/nephele/imgsvr/proc"
	"github.com/ctripcorp/nephele/imgsvr/storage"
	"github.com/ctripcorp/nephele/util"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var fdfsUrl = util.RegexpExt{regexp.MustCompile("fd/([a-zA-Z]+)/(.*)")}
var nfs1Url = util.RegexpExt{regexp.MustCompile("t1/([a-zA-Z]+)/(.*)")}
var nfs2Url = util.RegexpExt{regexp.MustCompile("([a-zA-Z]+)/(.*)")}

const fd = "fd"
const nfs1 = "nfs1"
const nfs2 = "nfs2"
const nfs = "nfs"
const Image = "Image"

//var StartPort int

type imgHandle struct {
	inImg  *img4g.Image
	chain  *proc.ProcessorChain
	status chan bool
}

var imgHandleChan = make(chan imgHandle, 1000)

//sourceType, channel, path
func ParseUri(path string) (string, string, string) {
	var sourceType = fd
	params, ok := fdfsUrl.FindStringSubmatchMap(path)
	if !ok {
		sourceType = nfs1
		params, ok = nfs1Url.FindStringSubmatchMap(path)
		if !ok {
			sourceType = nfs2
			params, ok = nfs2Url.FindStringSubmatchMap(path)
			if !ok {
				sourceType = ""
			}
		}
	}

	channel := params[":1"]
	p := params[":2"]
	switch sourceType {
	case fd:
		return sourceType, strings.ToLower(channel), p
	case nfs1:
		return nfs, strings.ToLower(channel), getTargetDir(channel, nfs1) + channel + "/" + p
	case nfs2:
		return nfs, strings.ToLower(channel), getTargetDir(channel, nfs1) + channel + "/" + p
	}
	return sourceType, "", ""
}

func getTargetDir(channel, storagetype string) string {
	dir, _ := data.GetDirPath(channel, storagetype)
	return dir
}

func JoinString(args ...string) string {
	var buf bytes.Buffer
	for _, v := range args {
		buf.WriteString(v)
	}
	return buf.String()
}

func GetStorage(storageType string, path string) (storage.Storage, error) {
	var srg storage.Storage
	switch storageType {
	case fd:
		domain, err := data.GetFdfsDomain()
		if err != nil {
			return nil, err
		}
		port := data.GetFdfsPort()
		srg = &storage.Fdfs{
			Path:          path,
			TrackerDomain: domain,
			Port:          port,
		}
	case nfs:
		srg = &storage.Nfs{path}
	}
	if srg == nil {
		return nil, errors.New("storageType(" + storageType + ") isn't supported!")
	}
	return srg, nil
}
func GetImage(storageType string, path string) ([]byte, error) {
	srg, err := GetStorage(storageType, path)
	if err != nil {
		return nil, err
	}
	return srg.GetImage()
}

func GetIP() string {
	ifs, err := net.Interfaces()
	if err != nil {
		return ""
	}
	if len(ifs) < 1 {
		return ""
	}
	ifi, err := net.InterfaceByName(ifs[0].Name)
	if err != nil {
		return ""
	}
	addrs, err := ifi.Addrs()
	if err != nil {
		return ""
	}
	if len(addrs) < 1 {
		return ""
	}
	return addrs[0].String()
}

func GetHttp(url string) ([]byte, error) {
	timeout := time.Duration(time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	rsp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	bts, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	return bts, nil
}

func PostHttp(uri string, data url.Values) ([]byte, error) {
	timeout := time.Duration(time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	rsp, err := client.PostForm(uri, data)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	bts, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	return bts, nil
}

func KillProcessByPort(port string) error {
	cmd := exec.Command("sh", "-c", JoinString("lsof -i:", port, "|grep LISTEN|awk '{print $2}'"))
	//	cmd.Env = append(cmd.Env, os.Getenv("PATH"))
	bts, err := cmd.Output()
	if err != nil {
		return err
	}
	pid := strings.TrimSpace(string(bts))
	if pid == "" {
		return nil
	}
	id, err := strconv.Atoi(pid)
	if err != nil {
		return err
	}
	p, err := os.FindProcess(id)
	if err != nil {
		return err
	}
	return p.Kill()
}

// Alloc        uint64      bytes allocated and still in use // 已分配且仍在使用的字节数
// 	TotalAlloc   uint64      // bytes allocated (even if freed) // 已分配（包括已释放的）字节数
// 	Sys          uint64      // bytes obtained from system (sum of XxxSys below) // 从系统中获取的字节数（应当为下面 XxxSys 之和）
// 	Mallocs      uint64      // number of mallocs // malloc 数
// 	Frees        uint64      // number of frees // free 数
// 	HeapAlloc    uint64      // bytes allocated and still in use // 已分配且仍在使用的字节数
// 	HeapSys      uint64      // bytes obtained from system // 从系统中获取的字节数
// 	HeapIdle     uint64      // bytes in idle spans // 空闲区间的字节数
// 	HeapInuse    uint64      // bytes in non-idle span // 非空闲区间的字节数
// 	HeapReleased uint64      // bytes released to the OS // 释放给OS的字节数
// 	HeapObjects  uint64      // total number of allocated objects// 已分配对象的总数
// 	OtherSys     uint64      // other system allocations // 其它系统分配
// 	NextGC       uint64      // next run in HeapAlloc time (bytes) // 下次运行的 HeapAlloc 时间（字节）
// 	LastGC       uint64      // last run in absolute time (ns) // 上次运行的绝对时间（纳秒 ns）
// 	PauseNs      [256]uint64 // circular buffer of recent GC pause times, most recent at [(NumGC+255)%256]
// 	NumGC        uint32

func GetStats() map[string]string {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	var second uint64 = 1000000000
	var memory uint64 = 1024 * 1024
	return map[string]string{"Alloc": strconv.FormatUint(mem.Alloc/memory, 10),
		"TotalAlloc":   strconv.FormatUint(mem.TotalAlloc/memory, 10),
		"Sys":          strconv.FormatUint(mem.Sys/memory, 10),
		"Mallocs":      strconv.FormatUint(mem.Mallocs, 10),
		"Frees":        strconv.FormatUint(mem.Frees, 10),
		"HeapAlloc":    strconv.FormatUint(mem.HeapAlloc/memory, 10),
		"HeapSys":      strconv.FormatUint(mem.HeapSys/memory, 10),
		"HeapIdle":     strconv.FormatUint(mem.HeapIdle/memory, 10),
		"HeapInuse":    strconv.FormatUint(mem.HeapInuse/memory, 10),
		"HeapReleased": strconv.FormatUint(mem.HeapReleased/memory, 10),
		"HeapObjects":  strconv.FormatUint(mem.HeapObjects, 10),
		"OtherSys":     strconv.FormatUint(mem.OtherSys, 10),
		"NextGC":       strconv.FormatUint(mem.NextGC/second, 10),
		"LastGC":       strconv.FormatUint(mem.LastGC/second, 10),
		"PauseNs":      strconv.FormatUint(mem.PauseNs[(mem.NumGC+255)%256]/second, 10),
		"NumGC":        strconv.Itoa(int(mem.NumGC)),
	}
}
