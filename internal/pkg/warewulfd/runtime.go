package warewulfd

import (
    //"io"
    "net/http"
    //"os"
    "strconv"
	"strings"
    "encoding/json"

	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
)

func RuntimeOverlaySend(w http.ResponseWriter, req *http.Request) {
    daemonLogf("DEBUG: req.Method: %s\n", req.Method)
    var res map[string]interface{}

    switch req.Method {
    case "GET":
        daemonLogf("WARNING: legacy client using 'GET'.\n")
    case "POST":
        daemonLogf("DEBUG: new client using 'POST'.\n")
        err := json.NewDecoder(req.Body).Decode(&res)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
    default:
        daemonLogf("Sorry, only GET and POST methods are supported.")
    }

	conf, err := warewulfconf.New()
	if err != nil {
		daemonLogf("ERROR: Could not read Warewulf configuration file: %s\n", err)
		w.WriteHeader(503)
		return
	}

	nodes, err := node.New()
	if err != nil {
		daemonLogf("ERROR: Could not read node configuration file: %s\n", err)
		w.WriteHeader(503)
		return
	}

	remote := strings.Split(req.RemoteAddr, ":")
	port, err := strconv.Atoi(remote[1])
	if err != nil {
		daemonLogf("ERROR: Could not convert port to integer: %s\n", remote[1])
		w.WriteHeader(503)
		return
	}

	if err != nil {
		daemonLogf("ERROR: Could not load configuration file: %s\n", err)
		return
	}

	if conf.Warewulf.Secure {
		if port >= 1024 {
			daemonLogf("DENIED: Connection coming from non-privledged port: %s\n", req.RemoteAddr)
			w.WriteHeader(401)
			return
		}
	}

    var n node.NodeInfo
    if !conf.Warewulf.MacIdentify {
        n, err = nodes.FindByIpaddr(remote[0])
        if err != nil {
            daemonLogf("WARNING: Could not find node by IP address: %s\n", remote[0])
            w.WriteHeader(404)
            return
        }
    } else {
        for _, hwa := range(res){
            n, err = nodes.FindByHwaddr(hwa.(string))
            if n.Id.Defined(){
                daemonLogf("DEBUG: nodeId: %s, HardwareAddr: %s\n", n.Id.Get(), hwa)
                break
            }
        }
    }

	if !n.Id.Defined() {
		daemonLogf("REQ:   %15s: %s (unknown/unconfigured node)\n", n.Id.Get(), req.URL.Path)
		w.WriteHeader(404)
		return
	} else {
		daemonLogf("REQ:   %15s: %s\n", n.Id.Get(), req.URL.Path)
	}

	if n.RuntimeOverlay.Defined() {
		fileName := config.RuntimeOverlayImage(n.Id.Get())

		if conf.Warewulf.AutobuildOverlays {
			if !util.IsFile(fileName) || util.PathIsNewer(fileName, node.ConfigFile) || util.PathIsNewer(fileName, config.RuntimeOverlaySource(n.RuntimeOverlay.Get())) {
				daemonLogf("BUILD: %15s: Runtime Overlay\n", n.Id.Get())
				_ = overlay.BuildRuntimeOverlay([]node.NodeInfo{n})
			}
		}

		//err := sendFile(w, fileName, n.Id.Get())
        //err := sendRuntimeOverlay(w, fileName)
        http.ServeFile(w, req, fileName)
		if err != nil {
			daemonLogf("ERROR: %s\n", err)
		}
	} else {
		w.WriteHeader(503)
		daemonLogf("WARNING: No 'runtime system-overlay' set for node %s\n", n.Id.Get())
	}
}

//func sendRuntimeOverlay(w http.ResponseWriter, filename string) error {
//    fd, err := os.Open(filename)
//    defer fd.Close()
//    if err != nil {
//        http.Error(w, err.Error(), http.StatusInternalServerError)
//        //return
//    }
//
//
//    if _, err := io.Copy(w, fd); err != nil {
//        http.Error(w, err.Error(), http.StatusInternalServerError)
//        //return
//    }
//    return nil
//}
