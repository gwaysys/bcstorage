package server

import (
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gwaylib/errors"
	"github.com/gwaylib/log"
)

func init() {
	RegisterFileHandle("/file/write", writeHandler)
}

func writeHandler(w http.ResponseWriter, r *http.Request) error {
	authedTarget, _, err := _fileHandler.authWrite(r)
	if err != nil {
		return writeMsg(w, 401, errors.As(err).Code())
	}

	to := authedTarget
	posStr := r.FormValue("pos")
	pos, _ := strconv.ParseInt(posStr, 10, 64)
	dir := filepath.Dir(to)
	// TODO: confirm file perm
	if err := os.MkdirAll(dir, 0755); err != nil {
		return writeMsg(w, 500, err.Error())
	}
	toFile, err := os.OpenFile(to, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return writeMsg(w, 403, errors.As(err, to).Error())
	}
	defer toFile.Close()

	toStat, err := toFile.Stat()
	if err != nil {
		return writeMsg(w, 500, errors.As(err).Error())
	}
	now := time.Now()
	if now.Sub(toStat.ModTime()) > 24*time.Hour {
		return writeMsg(w, 403, "file has been locked, need delete manually for rewritten")
	}

	if _, err := toFile.Seek(pos, 0); err != nil {
		return writeMsg(w, 500, errors.As(err).Error())
	}

	var uploader io.ReadCloser
	var size = int64(0)
	mFile, hFile, err := r.FormFile("file")
	if err != nil {
		uploader = r.Body
		size = r.ContentLength
	} else {
		uploader = mFile
		size = hFile.Size
	}
	log.Infof("Upload file %s, offset:%d, size:%d, from %s", to, pos, size, r.RemoteAddr)

	written, err := io.Copy(toFile, uploader)
	if err != nil {
		toFile.Close()
		log.Warn(errors.As(err))
		return writeMsg(w, 500, errors.As(err).Code())
	}
	// flush the data?
	toFile.Close()
	uploader.Close()

	if written < size {
		return writeMsg(w, 201, fmt.Sprintf("upload size(%d) not match written size(%d)", size, written))
	}

	msg := "success"
	switch r.FormValue("checksum") {
	case "sha1":
		toFile, err = os.Open(to)
		if err != nil {
			return writeMsg(w, 403, errors.As(err, to).Error())
		}
		if _, err := toFile.Seek(0, 0); err != nil {
			return writeMsg(w, 500, errors.As(err).Error())
		}
		ah := sha1.New()
		if _, err := io.Copy(ah, toFile); err != nil {
			return writeMsg(w, 500, errors.As(err).Error())
		}
		aSum := ah.Sum(nil)
		msg = fmt.Sprintf("%x", aSum)
	}

	// TODO: generate file id
	return writeMsg(w, 200, msg)
}
