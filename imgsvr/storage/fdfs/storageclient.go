package fdfs

import (
	"errors"
	"fmt"
	"github.com/ctripcorp/ghost/pool"
	"net"
	"os"
)

const (
	STORAGE_MIN_CONN int = 5
	STORAGE_MAX_CONN int = 5
)

type storageClient struct {
	*poolInfo
	pool.Pool
}

func newStorageClient(host string, port int) (*storageClient, error) {
	pInfo := &poolInfo{host, port, STORAGE_MIN_CONN, STORAGE_MAX_CONN}
	p, err := pInfo.newPool()
	if err != nil {
		return nil, err
	}
	return &storageClient{pInfo, p}, nil

}

func (this *storageClient) storageDownloadToFile(
	storeInfo *storageInfo, localFilename string, offset int64,
	downloadSize int64, fileName string) (*downloadRsp, error) {
	return this.storageDownload(storeInfo, localFilename, offset, downloadSize, FDFS_DOWNLOAD_TO_FILE, fileName)
}

func (this *storageClient) storageDownloadToBuffer(
	storeInfo *storageInfo, fileBuffer []byte, offset int64,
	downloadSize int64, fileName string) (*downloadRsp, error) {
	return this.storageDownload(storeInfo, fileBuffer, offset, downloadSize, FDFS_DOWNLOAD_TO_BUFFER, fileName)
}

func (this *storageClient) storageDownload(storeInfo *storageInfo, fileContent interface{}, offset int64, downloadSize int64, downloadType int, fileName string) (*downloadRsp, error) {
	var (
		conn          net.Conn
		reqBuf        []byte
		recvBuff      []byte
		localFilename string
		recvSize      int64
		err           error
	)
	//get a connetion from pool
	conn, err = this.Get()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	//request header
	rh := &reqHeader{}
	rh.cmd = STORAGE_PROTO_CMD_DOWNLOAD_FILE
	rh.pkgLen = int64(FDFS_PROTO_PKG_LEN_SIZE*2 + FDFS_GROUP_NAME_MAX_LEN + len(fileName))
	if err = rh.send(conn); err != nil {
		return nil, err
	}

	//request body
	req := &downloadReq{}
	req.offset = offset
	if downloadSize > 0 {
		req.downloadSize = downloadSize
	}
	req.groupName = storeInfo.groupName
	req.fileName = fileName
	reqBuf = req.marshal()
	if err = tcpSendData(conn, reqBuf); err != nil {
		return nil, err
	}

	//receive header
	if err = rh.recv(conn); err != nil {
		return nil, err
	}
	if rh.status != 0 {
		return nil, Errno{int(rh.status)}
	}

	//receive body
	switch downloadType {
	case FDFS_DOWNLOAD_TO_FILE:
		if localFilename, ok := fileContent.(string); ok {
			recvSize, err = tcpRecvFile(conn, localFilename, rh.pkgLen)
		}
	case FDFS_DOWNLOAD_TO_BUFFER:
		if _, ok := fileContent.([]byte); ok {
			recvBuff, recvSize, err = tcpRecvResponse(conn, rh.pkgLen)
		}
	}
	if err != nil {
		return nil, err
	}

	if downloadSize > 0 && recvSize < downloadSize {
		errmsg := "[-] Error: Storage response length is not match, "
		errmsg += fmt.Sprintf("expect: %d, actual: %d", rh.pkgLen, recvSize)
		return nil, errors.New(errmsg)
	}

	dr := &downloadRsp{}
	dr.fileId = storeInfo.groupName + string(os.PathSeparator) + fileName
	if downloadType == FDFS_DOWNLOAD_TO_FILE {
		dr.content = localFilename
	} else {
		dr.content = recvBuff
	}
	dr.downloadSize = recvSize
	return dr, nil
}
