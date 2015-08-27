package fdfs

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/ctripcorp/nephele/Godeps/_workspace/src/github.com/ctripcorp/ghost/pool"
	"net"
	"time"
)

const (
	TRACKER_MIN_CONN        int           = 5
	TRACKER_MAX_CONN        int           = 5
	TRACKER_MAX_IDLE        time.Duration = 119 * time.Second
	TRACKER_NETWORK_TIMEOUT time.Duration = 10 * time.Second
)

type trackerClient struct {
	host string
	port int
	pool.Pool
}

func newTrackerClient(host string, port int) (*trackerClient, error) {
	client := &trackerClient{host: host, port: port}
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
	//get a connection from pool
	conn, err := this.Get()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	buffer := new(bytes.Buffer)
	//package length
	binary.Write(buffer, binary.BigEndian, int64(FDFS_GROUP_NAME_MAX_LEN+len(fileName)))
	//cmd
	buffer.WriteByte(byte(cmd))
	//status
	buffer.WriteByte(byte(0))
	//16 bit groupName
	groupNameBytes := bytes.NewBufferString(groupName).Bytes()
	for i := 0; i < 15; i++ {
		if i >= len(groupNameBytes) {
			buffer.WriteByte(byte(0))
		} else {
			buffer.WriteByte(groupNameBytes[i])
		}
	}
	buffer.WriteByte(byte(0))
	// fileName
	fileNameBytes := bytes.NewBufferString(fileName).Bytes()
	for i := 0; i < len(fileNameBytes); i++ {
		buffer.WriteByte(fileNameBytes[i])
	}
	//send request
	if err := tcpSend(conn, buffer.Bytes(), TRACKER_NETWORK_TIMEOUT); err != nil {
		errMsg := fmt.Sprintf("send to tracker server %v fail, error info: %v", conn.RemoteAddr().String(), err.Error())
		return nil, errors.New(errMsg)
	}
	//receive body
	recvBuff, err := recvResponse(conn, TRACKER_NETWORK_TIMEOUT)
	if err != nil {
		return nil, err
	}
	buff := bytes.NewBuffer(recvBuff)
	// #recv_fmt |-group_name(16)-ipaddr(16-1)-port(8)-store_path_index(1)|
	groupName, err = readCstr(buff, FDFS_GROUP_NAME_MAX_LEN)
	ipAddr, err := readCstr(buff, IP_ADDRESS_SIZE-1)
	var port int64
	binary.Read(buff, binary.BigEndian, &port)
	var storePathIndex uint8
	binary.Read(buff, binary.BigEndian, &storePathIndex)
	return &storageInfo{ipAddr, int(port), groupName, int(storePathIndex)}, nil
}

//factory method used for dial
func (this *trackerClient) makeConn() (net.Conn, error) {
	addr := fmt.Sprintf("%s:%d", this.host, this.port)
	event := globalCat.NewEvent("DialTracker", addr)
	defer func() {
		event.Complete()
	}()
	conn, err := net.DialTimeout("tcp", addr, TRACKER_NETWORK_TIMEOUT)
	if err != nil {
		errMsg := fmt.Sprintf("dial tracker %v fail, error info: %v", addr, err.Error())
		event.AddData("detail", errMsg)
		event.SetStatus("ERROR")
		return nil, errors.New(errMsg)
	}
	event.SetStatus("0")
	return conn, nil
}
