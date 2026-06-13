package server

import (
	"net/http"
	"os"

	"github.com/gwaylib/errors"
	"github.com/gwaylib/log"
)

func init() {
	RegisterFileHandle("/file/move", moveHandler)
}

func moveHandler(w http.ResponseWriter, r *http.Request) error {
	authedTarget, fAuth, err := _fileHandler.authWrite(r)
	if err != nil {
		return writeMsg(w, 401, errors.As(err).Code())
	}

	oldName := authedTarget
	target, ok := fAuth.InAuthPath(r.FormValue("new"))
	if !ok {
		return writeMsg(w, 403, "new path is outside authed")
	}
	if err := os.Rename(oldName, target); err != nil {
		return errors.As(err)
	}
	log.Infof("Rename file %s to %s, from %s", oldName, target, r.RemoteAddr)
	return nil
}
