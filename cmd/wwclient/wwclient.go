package main

import (
	"bytes"
	"encoding/json"
	//"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	//"net/url"
	"os"
	"os/exec"
	"time"

	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

func main() {
	if os.Args[0] == "/warewulf/bin/wwclient" {
		err := os.Chdir("/")
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "failed to change dir: %s", err)
			os.Exit(1)
		}
		log.Printf("Updating live file system LIVE, cancel now if this is in error")
		time.Sleep(5000 * time.Millisecond)
	} else {
		fmt.Printf("Called via: %s\n", os.Args[0])
		fmt.Printf("Runtime overlay is being put in '/warewulf/wwclient-test' rather than '/'\n")
		err := os.MkdirAll("/warewulf/wwclient-test", 0755)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "failed to create dir: %s", err)
			os.Exit(1)
		}

		err = os.Chdir("/warewulf/wwclient-test")
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "failed to change dir: %s", err)
			os.Exit(1)
		}
	}

	conf, err := warewulfconf.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not get Warewulf configuration: %s\n", err)
		os.Exit(1)
	}

	localTCPAddr := net.TCPAddr{}
	if conf.Warewulf.Secure {
		// Setup local port to something privileged (<1024)
		localTCPAddr.Port = 987
	} else {
		fmt.Printf("INFO: Running from an insecure port\n")
	}

	webclient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				LocalAddr: &localTCPAddr,
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	// get all network interfaces and create json payload
	interfaces, _ := net.Interfaces()
	nics := make(map[string]string)
	for _, inter := range interfaces {
		nics[inter.Name] = inter.HardwareAddr.String()
	}
	hwAddrs, err := json.Marshal(nics)
	if err != nil {
		panic(err)
	}

	for {
		var resp *http.Response
		counter := 0

		for {
			var err error

			getString := fmt.Sprintf("http://%s:%d/overlay-runtime/", conf.Ipaddr, conf.Warewulf.Port)
			resp, err = webclient.Post(getString, "application/json", bytes.NewBuffer(hwAddrs))
			if err == nil {
				break
			} else {
				if counter > 60 {
					counter = 0
				}
				if counter == 0 {
					log.Println(err)
				}
				counter++
			}
			time.Sleep(1000 * time.Millisecond)
		}

		if resp.StatusCode != 200 {
			log.Printf("Not updating runtime overlay, got status code: %d\n", resp.StatusCode)
			time.Sleep(60000 * time.Millisecond)
			continue
		}

		log.Printf("Updating system\n")
		command := exec.Command("/bin/sh", "-c", "gzip -dc | cpio -iu")
		command.Stdin = resp.Body
		err := command.Run()
		if err != nil {
			log.Printf("ERROR: Failed running CPIO: %s\n", err)
		}

		if conf.Warewulf.UpdateInterval > 0 {
			time.Sleep(time.Duration(conf.Warewulf.UpdateInterval*1000) * time.Millisecond)
		} else {
			time.Sleep(30000 * time.Millisecond * 1000)
		}
	}
}
