package server

import (
	"net/http"

	"github.com/gwaylib/errors"
)

func init() {
	RegisterFileHandle("/file/read", readHandler)
}

func readHandler(w http.ResponseWriter, r *http.Request) error {
	authedTarget, _, err := _fileHandler.authRead(r)
	if err != nil {
		return writeMsg(w, 401, errors.As(err).Code())
	}
	http.ServeFile(w, r, authedTarget)
	return nil
}
