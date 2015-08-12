package fdfs

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

type Errno struct {
	status int
}

func (e Errno) Error() string {
	errmsg := fmt.Sprintf("errno [%d] ", e.status)
	switch e.status {
	case 17:
		errmsg += "File Exist"
	case 22:
		errmsg += "Argument Invlid"
	}
	return errmsg
}

type FdfsConfigParser struct{}

func fdfsCheckFile(filename string) error {
	if _, err := os.Stat(filename); err != nil {
		return err
	}
	return nil
}

func readCstr(buff io.Reader, length int) (string, error) {
	str := make([]byte, length)
	n, err := buff.Read(str)
	if err != nil || n != len(str) {
		return "", Errno{255}
	}

	for i, v := range str {
		if v == 0 {
			str = str[0:i]
			break
		}
	}
	return string(str), nil
}
func getFileExt(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) >= 2 {
		return parts[len(parts)-1]
	}
	return ""
}

func splitRemoteFileId(remoteFileId string) ([]string, error) {
	parts := strings.SplitN(remoteFileId, "/", 2)
	if len(parts) < 2 {
		return nil, errors.New("error remoteFileId")
	}
	return parts, nil
}

func tcpSend(conn net.Conn, bytesStream []byte, timeout time.Duration) error {
	if err := conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	if _, err := conn.Write(bytesStream); err != nil {
		return err
	}
	return nil
}

func tcpRecv(conn net.Conn, bufferSize int64, timeout time.Duration) ([]byte, error) {
	buff := make([]byte, bufferSize)
	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}
	if _, err := io.ReadFull(conn, buff); err != nil {
		return nil, err
	}
	return buff, nil
}
