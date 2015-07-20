package imgsvr

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

func ModifyNginxconf(path string, listenPort string, ports map[string]int) error {
	if path == "" {
		return nil
	}
	if ports == nil || len(ports) < 1 {
		return nil
	}
	bts, err := ioutil.ReadFile(path + "conf/nginx.conf")
	if err != nil {
		return err
	}
	content := string(bts)
	arr := strings.Split(content, "\n")
	newarr := []string{}

	var hasGohttp bool = false
	var hasGohttpServer = false
	var tmp string = arr[0]

	for _, v := range arr {
		if strings.TrimSpace(v) == "" {
			continue
		}
		if strings.Contains(tmp, "upstream go_http") && strings.Contains(v, "127.0.0.1:") {
			continue
		}
		newarr = append(newarr, tmp)
		if strings.Contains(tmp, "upstream go_http") {
			t := strings.TrimSpace(tmp)
			if t[0] != '#' {
				hasGohttp = true
				for p, _ := range ports {
					if p != "" {
						newarr = append(newarr, " server 127.0.0.1:"+p+";")
					}
				}
			}
		}

		tmp = v

		if strings.Contains(v, "http://go_http") {
			t := strings.TrimSpace(v)
			if t[0] != '#' {
				hasGohttpServer = true
			}
		}
	}
	if !hasGohttp {
		newarr = append(newarr, " upstream go_http{")
		for p, _ := range ports {
			if p != "" {
				newarr = append(newarr, " server 127.0.0.1:"+p+";")
			}
		}
		newarr = append(newarr, "keepalive 700;")
		newarr = append(newarr, "}")
	}

	if !hasGohttpServer {
		if listenPort == "" {
			listenPort = "80"
		}
		newarr = append(newarr, " server {", JoinString(" listen "+listenPort+";"), " server_name go_http;", " access_log off;", " error_log /dev/null crit;", " location / {", " proxy_pass http://go_http;", " proxy_http_version 1.1;", " proxy_set_header Connection \"\"", "}", "}")
	}

	newarr = append(newarr, tmp)
	s := strings.Join(newarr, "\n")
	return ioutil.WriteFile(path+"conf/nginx.conf", []byte(s), os.ModeType)
}

func RestartNginx(path string) error {
	tmp := path + "sbin/nginx"
	cmd := exec.Command(tmp, "-s", "reload")
	err := cmd.Start()
	if err != nil {
		return err
	}
	return nil
}
