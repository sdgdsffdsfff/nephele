package fdfs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/ctripcorp/cat"
	"github.com/ctripcorp/ghost/pool"
	"net"
	"time"
)

const (
	TRACKER_MIN_CONN int           = 5
	TRACKER_MAX_CONN int           = 5
	TRACKER_MAX_IDLE time.Duration = 10 * time.Minute
)

type trackerClient struct {
	host string
	port int
	pool.Pool
}

func newTrackerClient(host string, port int) (*trackerClient, error) {
	client := &trackerClient{host:host, port:port}
	p, err := pool.NewBlockingPool(TRACKER_MIN_CONN, TRACKER_MAX_CONN, TRACKER_MAX_IDLE, client.makeConn)
	if err != nil {
		return nil, err
	}
	client.Pool = p
	return client, nil
}

//fetch a  download stroage from tracker
func (this *trackerClient) trackerQueryStorageFetch(groupName string, fileName string) (*storageInfo, error) {
	return this.trackerQueryStorage(groupName, fileName, TRACKER_PROTO_CMD_SERVICE_QUERY_FETCH_ONE)
}

//query stroage sever with specific command
func (this *trackerClient) trackerQueryStorage(groupName string, fileName string, cmd int8) (*storageInfo, error) {
	var (
		conn     net.Conn
		recvBuff []byte
		err      error
	)
	//get a connection from pool
	conn, err = this.Get()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	rh := &reqHeader{}
	rh.pkgLen = int64(FDFS_GROUP_NAME_MAX_LEN + len(fileName))
	rh.cmd = cmd
	if err := rh.send(conn); err != nil {
		return nil, err
	}

	// #query_fmt: |-group_name(16)-filename(file_name_len)-|
	queryBuffer := new(bytes.Buffer)
	// 16 bit groupName
	groupNameBytes := bytes.NewBufferString(groupName).Bytes()
	for i := 0; i < 16; i++ {
		if i >= len(groupNameBytes) {
			queryBuffer.WriteByte(byte(0))
		} else {
			queryBuffer.WriteByte(groupNameBytes[i])
		}
	}
	// fileNameLen bit fileName
	fileNameBytes := bytes.NewBufferString(fileName).Bytes()
	for i := 0; i < len(fileNameBytes); i++ {
		queryBuffer.WriteByte(fileNameBytes[i])
	}
	if err := tcpSendData(conn, queryBuffer.Bytes()); err != nil {
		return nil, err
	}

	//response header
	if err := rh.recv(conn); err != nil {
		return nil, err
	}
	if rh.status != 0 {
		return nil, Errno{int(rh.status)}
	}

	var (
		ipAddr         string
		port           int64
		storePathIndex uint8
	)
	recvBuff, _, err = tcpRecvResponse(conn, rh.pkgLen)
	if err != nil {
		return nil, err
	}
	buff := bytes.NewBuffer(recvBuff)
	// #recv_fmt |-group_name(16)-ipaddr(16-1)-port(8)-store_path_index(1)|
	groupName, err = readCstr(buff, FDFS_GROUP_NAME_MAX_LEN)
	ipAddr, err = readCstr(buff, IP_ADDRESS_SIZE-1)
	binary.Read(buff, binary.BigEndian, &port)
	binary.Read(buff, binary.BigEndian, &storePathIndex)
	return &storageInfo{ipAddr, int(port), groupName, int(storePathIndex)}, nil
}

func (this *trackerClient) makeConn() (net.Conn, error) {
	Cat := cat.Instance()
	addr := fmt.Sprintf("%s:%d", this.host, this.port)
	conn, err := net.DialTimeout("tcp", addr, 30*time.Second)
	if err != nil {
		event := Cat.NewEvent("DialTracker", "Fail")
		event.AddData("addr", addr)
		event.AddData("detail", err.Error())
		event.SetStatus("ERROR")
		event.Complete()
		return nil, err
	}
	event := Cat.NewEvent("DialTracker", "Success")
	event.AddData("addr", addr)
	event.SetStatus("0")
	event.Complete()
	return conn, nil
}
