package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/gwaylib/errors"
	"github.com/gwaylib/log"
)

func init() {
	// TODO: rollback
	RegisterFileHandle("/file/delete", deleteHandler)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) error {
	authedTarget, fAuth, err := _fileHandler.authWrite(r)
	if err != nil {
		return writeMsg(w, 401, errors.As(err).Code())
	}

	oldFile := authedTarget
	bakName := "." + uuid.New().String()

	bakKey := fmt.Sprintf(_leveldb_prefix_del, fAuth._space, bakName)
	if err := PutLevelDB(bakKey, oldFile); err != nil {
		return errors.As(err)
	}

	bakFile := filepath.Join(fAuth.absSpace, bakName)
	if err := os.Rename(oldFile, bakFile); err != nil {
		// if err := os.Remove(path); err != nil {
		if !os.IsNotExist(err) {
			return errors.As(err)
		}
	}
	log.Warnf("Delete file:%s,from:%s", oldFile, r.RemoteAddr)
	return nil
}
