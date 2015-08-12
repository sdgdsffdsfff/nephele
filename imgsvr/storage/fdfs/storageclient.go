package fdfs

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/ctripcorp/ghost/pool"
	"net"
	"time"
)

const (
	STORAGE_MIN_CONN        int           = 5
	STORAGE_MAX_CONN        int           = 5
	STORAGE_MAX_IDLE        time.Duration = 10 * time.Hour
	STORAGE_NETWORK_TIMEOUT time.Duration = 10 * time.Second
)

type storageClient struct {
	host string
	port int
	pool.Pool
}

func newStorageClient(host string, port int) (*storageClient, error) {
	client := &storageClient{host: host, port: port}
	p, err := pool.NewBlockingPool(STORAGE_MIN_CONN, STORAGE_MAX_CONN, STORAGE_MAX_IDLE, client.makeConn)
	if err != nil {
		return nil, err
	}
	client.Pool = p
	return client, nil

}

func (this *storageClient) storageDownload(storeInfo *storageInfo, offset int64, downloadSize int64, fileName string) ([]byte, error) {
	//get a connetion from pool
	conn, err := this.Get()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	//request header
	buffer := new(bytes.Buffer)
	//package length:file_offset(8)  download_bytes(8)  group_name(16)  file_name(n)
	binary.Write(buffer, binary.BigEndian, int64(32+len(fileName)))
	//cmd
	buffer.WriteByte(byte(STORAGE_PROTO_CMD_DOWNLOAD_FILE))
	//status
	buffer.WriteByte(byte(0))
	//offset
	binary.Write(buffer, binary.BigEndian, offset)
	//download bytes
	binary.Write(buffer, binary.BigEndian, downloadSize)
	//16 bit groupName
	groupNameBytes := bytes.NewBufferString(storeInfo.groupName).Bytes()
	for i := 0; i < 15; i++ {
		if i >= len(groupNameBytes) {
			buffer.WriteByte(byte(0))
		} else {
			buffer.WriteByte(groupNameBytes[i])
		}
	}
	buffer.WriteByte(byte(0))
	// fileNameLen bit fileName
	fileNameBytes := bytes.NewBufferString(fileName).Bytes()
	for i := 0; i < len(fileNameBytes); i++ {
		buffer.WriteByte(fileNameBytes[i])
	}
	//send request
	if err := tcpSend(conn, buffer.Bytes(), STORAGE_NETWORK_TIMEOUT); err != nil {
		return nil, errors.New(fmt.Sprintf("send to storage server %v fail, error info: %v", conn.RemoteAddr().String(), err.Error()))
	}
	//receive response header
	recvBuff, err := recvResponse(conn, STORAGE_NETWORK_TIMEOUT)
	if err != nil {
		return nil, err
		//try again
		//	if err = tcpSend(conn, buffer.Bytes(), STORAGE_NETWORK_TIMEOUT); err != nil {
		//		return nil, err
		//	}
		//	if recvBuff, err = recvResponse(conn, STORAGE_NETWORK_TIMEOUT); err != nil {
		//		return nil, err
		//	}
	}
	return recvBuff, nil
}

//factory method used to dial
func (this *storageClient) makeConn() (net.Conn, error) {
	addr := fmt.Sprintf("%s:%d", this.host, this.port)
	event := globalCat.NewEvent("DialStorage", addr)
	conn, err := net.DialTimeout("tcp", addr, STORAGE_NETWORK_TIMEOUT)
	if err != nil {
		errMsg := fmt.Sprintf("dial storage %v fail, error info: %v", addr, err.Error())
		event.AddData("detail", errMsg)
		event.SetStatus("ERROR")
		event.Complete()
		return nil, errors.New(errMsg)
	}
	event.SetStatus("0")
	event.Complete()
	return conn, nil
}
