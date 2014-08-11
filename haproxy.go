package main

import (
  "fmt"
	"os"
	"io/ioutil"
	"os/exec"
)


func generateHAProxyConfig(applicationMap map[string]Application) {
	tmp, err := ioutil.TempFile("", "haproxy.cfg")
	if err != nil {
		return
	}

	fmt.Fprintf(tmp, haproxyHeader)
	for appId, app := range applicationMap {
		fmt.Fprintf(tmp, "\nlisten %s\n  bind 0.0.0.0:%d\n  mode tcp\n  option tcplog\n  balance leastconn\n", appId, app.Ports[0])
		i := 0
		for _, task := range app.ApplicationInstances {
			fmt.Fprintf(tmp, "  server %s-%d %s:%d check\n", appId, i, task.Host, task.Ports[0])
			i++
		}
	}
	err = os.Rename(tmp.Name(), "/etc/haproxy/haproxy.cfg")
	if err != nil {
		fmt.Println("ERR Couldn't write /etc/haproxy/haproxy.cfg")
		fmt.Println(err)
		return
	}
	fmt.Println("INFO Written new /etc/haproxy/haproxy.cfg")
	cmd := exec.Command("service", "haproxy", "reload")
	err = cmd.Start()
	if err != nil {
		fmt.Println("ERR failed to reload HAProxy")
		return
	}
	err = cmd.Wait()
	if err != nil {
		fmt.Println("ERR failed to reload HAProxy")
		return
	}
}
