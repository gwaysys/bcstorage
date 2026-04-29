package main

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
	authPath, fAuth, err := authWrite(r)
	if err != nil {
		return writeMsg(w, 401, errors.As(err).Code())
	}

	oldName := authPath
	newName, err := validHttpFilePath(fAuth.spaceName, r.FormValue("new"))
	if err != nil {
		return writeMsg(w, 403, errors.As(err).Code())
	}
	if err := os.Rename(oldName, newName); err != nil {
		return errors.As(err)
	}
	log.Infof("Rename file %s to %s, from %s", oldName, newName, r.RemoteAddr)
	return nil
}
