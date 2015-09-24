package fdfs

import (
	"errors"
	"fmt"
	cat "github.com/ctripcorp/nephele/Godeps/_workspace/src/github.com/ctripcorp/cat.go"
	"github.com/ctripcorp/nephele/util"
	"math/rand"
	"strconv"
	"sync"
)

type FdfsClient interface {
	//first, the client get a connection from connetction pool, and connetct
	//to the tracker to query the ip and port of a download storage server
	//if the storage server doesn't exist, it will create a connection pool
	//and add it to the storge server pool map, otherwise it will get directly
	//from the pool map
	DownloadToBuffer(fileId string, catInstance cat.Cat) ([]byte, error)
}

//cat instance transferred by user
//var userCat cat.Cat
var globalCat cat.Cat

func init() {
	util.InitCat()
	globalCat = cat.Instance()
}

type fdfsClient struct {
	//tracker client containing a connetction pool
	tracker *trackerClient

	//storage client map
	storages map[string]*storageClient

	//use to read or write a storage client from map
	mutex sync.RWMutex
}

//NewFdfsClient create a connection pool to a tracker
//the tracker is selected randomly from tracker group
func NewFdfsClient(trackerHosts []string, trackerPort string) (FdfsClient, error) {
	//select a random tracker host from host list
	host := trackerHosts[rand.Intn(len(trackerHosts))]
	port, err := strconv.Atoi(trackerPort)
	if err != nil {
		return nil, err
	}
	tc, err := newTrackerClient(host, port)
	if err != nil {
		return nil, err
	}
	return &fdfsClient{tracker: tc, storages: make(map[string]*storageClient)}, nil
}

func (this *fdfsClient) DownloadToBuffer(fileId string, catInstance cat.Cat) ([]byte, error) {
	if catInstance == nil {
		return nil, errors.New("cat instance transferred to fdfs is nil")
	}
	buff, err := this.downloadToBufferByOffset(fileId, 0, 0, catInstance)
	if err != nil {
		return nil, err
	}
	return buff, nil
}

func (this *fdfsClient) downloadToBufferByOffset(fileId string, offset int64, downloadSize int64, catInstance cat.Cat) ([]byte, error) {
	//split file id to two parts: group name and file name
	tmp, err := splitRemoteFileId(fileId)
	if err != nil || len(tmp) != 2 {
		return nil, err
	}
	groupName := tmp[0]
	fileName := tmp[1]

	//query a download server from tracker
	storeInfo, err := this.tracker.trackerQueryStorageFetch(groupName, fileName)
	if err != nil {
		return nil, err
	}
	event := catInstance.NewEvent("ImgFromStorage", fmt.Sprintf("%s:%s", storeInfo.groupName, storeInfo.ipAddr))
	event.SetStatus("0")
	event.Complete()

	//get a storage client from storage map, if not exist, create a new storage client
	storeClient, err := this.getStorage(storeInfo.ipAddr, storeInfo.port)
	if err != nil {
		return nil, err
	}
	return storeClient.storageDownload(storeInfo, offset, downloadSize, fileName)
}

func (this *fdfsClient) getStorage(ip string, port int) (*storageClient, error) {
	storageKey := fmt.Sprintf("%s-%d", ip, port)
	//if the storage with the key exists, return the stroage
	//else create a new stroage and return
	if sc := this.queryStorage(storageKey); sc != nil {
		return sc, nil
	} else {
		this.mutex.Lock()
		defer this.mutex.Unlock()
		//reconfirm wheather the storage exists
		if sc, ok := this.storages[storageKey]; ok {
			return sc, nil
		} else {
			sc, err := newStorageClient(ip, port)
			if err != nil {
				return nil, err
			}
			this.storages[storageKey] = sc
			return sc, nil
		}
	}
}

//query a storage client from storage map by key
//if the storage not eixst, return nil
func (this *fdfsClient) queryStorage(key string) *storageClient {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	if sc, ok := this.storages[key]; ok {
		return sc
	} else {
		return nil
	}
}
