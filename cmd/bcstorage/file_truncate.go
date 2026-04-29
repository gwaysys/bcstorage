package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gwaylib/errors"
	"github.com/gwaylib/log"
)

func init() {
	RegisterFileHandle("/file/truncate", truncateHandler)
}

func truncateHandler(w http.ResponseWriter, r *http.Request) error {
	authPath, _, err := authWrite(r)
	if err != nil {
		return writeMsg(w, 401, errors.As(err).Code())
	}

	size, err := strconv.ParseInt(r.FormValue("size"), 10, 64)
	if err != nil {
		return writeMsg(w, 403, "file size failed")
	}

	if err := os.MkdirAll(filepath.Dir(authPath), 0755); err != nil {
		return errors.As(err, authPath)
	}

	f, err := os.OpenFile(authPath, os.O_RDWR|os.O_CREATE, 0644) // nolint
	if err != nil {
		return errors.As(err, authPath)
	}
	defer f.Close()

	log.Infof("Trucate %s, size:%d, from:%s", authPath, size, r.RemoteAddr)
	if err := f.Truncate(size); err != nil {
		return errors.As(err, authPath)
	}
	return nil
}
