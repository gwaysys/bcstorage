package server

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
	authedTarget, _, err := _fileHandler.authWrite(r)
	if err != nil {
		return writeMsg(w, 401, errors.As(err).Code())
	}

	size, err := strconv.ParseInt(r.FormValue("size"), 10, 64)
	if err != nil {
		return writeMsg(w, 403, "file size failed")
	}

	if err := os.MkdirAll(filepath.Dir(authedTarget), 0755); err != nil {
		return errors.As(err)
	}

	f, err := os.OpenFile(authedTarget, os.O_RDWR|os.O_CREATE, 0644) // nolint
	if err != nil {
		return errors.As(err)
	}
	defer f.Close()

	log.Infof("Trucate %s, size:%d, from:%s", authedTarget, size, r.RemoteAddr)
	if err := f.Truncate(size); err != nil {
		return errors.As(err)
	}
	return nil
}
