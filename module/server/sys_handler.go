package server

import (
	"net/http"
	"sync"
	"time"
)

type CheckCache struct {
	out        string
	createTime time.Time
}
type SysHandler struct {
	checkCache   *CheckCache
	checkCacheLk sync.Mutex
}

func (h *SysHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//log.Infof("from:%s,method:%s,path:%+v", r.RemoteAddr, r.Method, r.URL.Path)
	// route handler
	handle, ok := sysHandles[r.URL.Path]
	if !ok {
		writeMsg(w, 404, "Not found")
		return
	}

	if err := handle(w, r); err != nil {
		writeMsg(w, 500, err.Error())
		return
	}
	return
}

var sysHandles = map[string]HandleFunc{}

func RegisterSysHandle(path string, handle HandleFunc) {
	_, ok := sysHandles[path]
	if ok {
		panic("already registered:" + path)
	}
	sysHandles[path] = handle
}
