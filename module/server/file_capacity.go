package server

import (
	"net/http"
	"syscall"

	"github.com/gwaylib/errors"
)

func init() {
	RegisterFileHandle("/file/capacity", capacityHandler)
}

func capacityHandler(w http.ResponseWriter, r *http.Request) error {
	_, fAuth, err := _fileHandler.authWrite(r)
	if err != nil {
		return writeMsg(w, 401, errors.As(err).Code())
	}

	// implement the df -h
	fs := syscall.Statfs_t{}
	if err := syscall.Statfs(fAuth.absSpace, &fs); err != nil {
		return errors.As(err)
	}

	return writeJson(w, 200, fs)
}
