package server

import (
	"encoding/json/v2"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gwaylib/errors"
	"github.com/gwaylib/log"
)

type FileHandler struct {
	token   map[string]*FileToken
	tokenLk sync.Mutex

	gcQueue chan bool
	lastGc  time.Time
}

func NewFileHandler() *FileHandler {
	h := &FileHandler{
		token:   map[string]*FileToken{},
		tokenLk: sync.Mutex{},
		gcQueue: make(chan bool, 10),
	}
	go func() {
		for {
			<-h.gcQueue
			h.gcToken()
		}
	}()
	return h
}

func (h *FileHandler) gcToken() {
	now := time.Now()
	if now.Sub(h.lastGc) < time.Hour {
		return
	}

	// TODO: need lock? maybe user just delay the token
	gcTokens := []string{}
	if err := IterLevelDB(_leveldb_prefix_token, func(key, val []byte) error {
		fileToken := &FileToken{}
		if err := json.Unmarshal(val, fileToken); err != nil {
			return errors.As(err)
		}
		if fileToken.expiredAt.Unix() > 0 && fileToken.expiredAt.Sub(now) < 0 {
			// recycle expired tokens
			gcTokens = append(gcTokens, string(key[len("_token."):]))
		}
		return nil
	}); err != nil {
		log.Warn(errors.As(err))
		return
	}

	for _, token := range gcTokens {
		if err := h.DeleteToken(token); err != nil {
			log.Warn(errors.As(err))
			return
		}
	}
}

func (h *FileHandler) _AddGcEvent() {
	select {
	case h.gcQueue <- true:
	default:
		// queue is full
	}
}

func (h *FileHandler) GetToken(token string) (*FileToken, error) {
	h.tokenLk.Lock()
	fileToken, ok := h.token[token]
	h.tokenLk.Unlock()
	if ok {
		return fileToken, nil
	}
	fileToken = &FileToken{}
	if err := GetLevelDB(fmt.Sprintf(_leveldb_prefix_token, token), fileToken); err != nil {
		return nil, errors.As(err)
	}
	return fileToken, nil
}

// if the file is dir, the dir files will be auto grant
func (h *FileHandler) AddToken(userSpace, file, token, grant string, expAt time.Time) error {
	existToken, err := h.GetToken(token)
	if err != nil {
		if !errors.ErrNoData.Equal(err) {
			return errors.As(err)
		}
		// token not exist
	} else {
		if existToken._space != userSpace || existToken._file != file {
			return errors.New("token already used by other one")
		}
	}
	fileToken, err := NewFileToken(
		userSpace,
		file,
		grant,
		expAt,
	)
	if err != nil {
		return errors.As(err)
	}

	// add to leveldb
	if err := PutLevelDB(fmt.Sprintf(_leveldb_prefix_token, token), fileToken); err != nil {
		return errors.As(err)
	}

	// add to memory
	h.tokenLk.Lock()
	defer h.tokenLk.Unlock()
	h.token[token] = fileToken
	return nil
}

func (h *FileHandler) DeleteToken(token string) error {
	if err := DelLevelDB(fmt.Sprintf(_leveldb_prefix_token, token)); err != nil {
		return errors.As(err)
	}

	h.tokenLk.Lock()
	defer h.tokenLk.Unlock()
	delete(h.token, token)
	return nil
}

func (h *FileHandler) VerifyToken(file, token string) (string, *FileToken, error) {
	h._AddGcEvent()

	fileToken, err := h.GetToken(token)
	if err != nil {
		if errors.ErrNoData.Equal(err) {
			return "", nil, errors.New("auth not match")
		}
		return "", nil, errors.As(err)
	}

	now := time.Now()
	if fileToken.expiredAt.Unix() <= 0 && fileToken.grant == FILE_TOKEN_GRANT_ALL {
		// never expired
	} else if fileToken.expiredAt.Sub(now) < 0 {
		// expired
		return "", nil, errors.New("auth expired").As(fileToken.grant, fileToken.expiredAt.Unix())
	}

	if len(file) == 0 {
		return fileToken.AuthedPath(), fileToken, nil
	}

	tagetAbsPath, ok := fileToken.InAuthPath(file)
	if !ok {
		return "", nil, errors.New("unauth path")
	}
	return tagetAbsPath, fileToken, nil
}

func (h *FileHandler) authWrite(r *http.Request) (string, *FileToken, error) {
	token := r.FormValue("tk") // token
	file := r.FormValue("ta")  // target
	authedFile, fAuth, err := h.VerifyToken(file, token)
	if err != nil {
		return "", nil, errors.As(err)
	}
	if fAuth.ReadOnly() {
		return "", nil, errors.New("only read auth")
	}
	return authedFile, fAuth, nil
}

func (h *FileHandler) authRead(r *http.Request) (string, *FileToken, error) {
	token := r.FormValue("tk") // token
	file := r.FormValue("ta")  // target
	return h.VerifyToken(file, token)
}

// download file
// the publich dir files can read by every one
func (h *FileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//log.Infof("from:%s,method:%s,path:%+v", r.RemoteAddr, r.Method, r.URL.Path)

	// Cors
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	if r.Method == "OPTIONS" {
		writeMsg(w, 200, "OK")
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
}

var fileHandles = map[string]HandleFunc{}

func RegisterFileHandle(path string, handle HandleFunc) {
	_, ok := fileHandles[path]
	if ok {
		panic("already registered:" + path)
	}
	fileHandles[path] = handle
}
