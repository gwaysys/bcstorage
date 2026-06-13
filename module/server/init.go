package server

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gwaylib/errors"
)

var (
	_dataPath = "./data"
	_cfgPath  = "./data"

	_sysHandler  = &SysHandler{}
	_fileHandler = NewFileHandler()
)

func InitAll(dataPath, cfgPath string) error {
	if err := os.MkdirAll(dataPath, 0700); err != nil {
		return errors.As(err, dataPath)
	}
	if err := os.MkdirAll(cfgPath, 0700); err != nil {
		return errors.As(err, cfgPath)
	}
	_dataPath = dataPath
	_cfgPath = cfgPath

	// stop the chattr protect.
	// TODO: remove this code
	if err := ioutil.WriteFile(filepath.Join(_dataPath, "check.dat"), []byte("success"), 0600); err != nil {
		return errors.As(err, _dataPath)
	}
	if err := InitLevelDB(filepath.Join(_cfgPath, _leveldbPath)); err != nil {
		return errors.As(err, _cfgPath)
	}

	if err := initDaemonAuth(); err != nil {
		return errors.As(err)
	}
	if err := initMCPHandler(); err != nil {
		return errors.As(err)
	}
	return nil
}

func HttpFileListen(addr string) error {
	return http.ListenAndServe(addr, _fileHandler)
}

func HttpsAuthListen(addr, crtPath, keyPath string) error {
	return http.ListenAndServeTLS(addr, crtPath, keyPath, _sysHandler)
}
