package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gwaylib/errors"
	"github.com/gwaylib/log"
)

type FileToken struct {
	spaceName  string
	file       string
	createTime time.Time
}

type FileHandler struct {
	token   map[string]FileToken
	tokenLk sync.Mutex
}

func (h *FileHandler) gcToken() {
	now := time.Now()
	for key, val := range h.token {
		if now.Sub(val.createTime) > 3*time.Hour {
			delete(h.token, key)
		}
	}
}
func (h *FileHandler) AddToken(spaceName, file, token string) {
	h.tokenLk.Lock()
	defer h.tokenLk.Unlock()
	h.token[token] = FileToken{
		spaceName:  spaceName,
		file:       file,
		createTime: time.Now(),
	}
}
func (h *FileHandler) DelayToken(token string) bool {
	h.tokenLk.Lock()
	defer h.tokenLk.Unlock()
	t, ok := h.token[token]
	if !ok {
		return false
	}
	t.createTime = time.Now()
	h.token[token] = t
	return true
}

func (h *FileHandler) DeleteToken(token string) {
	h.tokenLk.Lock()
	defer h.tokenLk.Unlock()
	delete(h.token, token)
}

func (h *FileHandler) VerifyToken(space, file, token string) (FileToken, bool) {
	h.tokenLk.Lock()
	defer h.tokenLk.Unlock()
	h.gcToken()

	t, ok := h.token[token]
	if !ok {
		return FileToken{}, false
	}
	if t.spaceName != space || !strings.HasPrefix(file, t.file) {
		return FileToken{}, false
	}
	return t, true

}

func (h *FileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//log.Infof("from:%s,method:%s,path:%+v", r.RemoteAddr, r.Method, r.URL.Path)

	// Cors
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	if r.Method == "OPTIONS" {
		writeMsg(w, 200, "OK")
		return
	}

	// for public read
	if strings.HasPrefix(r.URL.Path, "/public") {
		//log.Infof("from:%s,method:%s,path:%+v", r.RemoteAddr, r.Method, r.URL.Path)
		paths := strings.Split(r.URL.Path, "/")
		if len(paths) < 4 {
			writeMsg(w, 404, "error paths")
			return
		}
		//log.Info("paths:", paths, paths[2])
		userSpace, ok := _userMap.GetSpace(paths[2])
		if !ok {
			writeMsg(w, 404, fmt.Sprintf("no userspace '%s'", paths[2]))
			return
		}
		if userSpace.Private {
			writeMsg(w, 401, "unauth")
			return
		}
		to, err := validHttpFilePath(paths[2], filepath.Join(paths[3:]...))
		if err != nil {
			//log.Info(errors.As(err))
			writeMsg(w, 404, errors.As(err).Code())
			return
		}

		log.Info("server file", to, r.URL.Path)

		http.ServeFile(w, r, to)
		return
	}

	// route handler
	handle, ok := fileHandles[r.URL.Path]
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

var fileHandles = map[string]HandleFunc{}

func RegisterFileHandle(path string, handle HandleFunc) {
	_, ok := fileHandles[path]
	if ok {
		panic("already registered:" + path)
	}
	fileHandles[path] = handle
}
