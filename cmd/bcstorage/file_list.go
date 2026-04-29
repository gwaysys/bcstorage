package main

import (
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gwaysys/bcstorage/lib/utils"
	"github.com/gwaylib/errors"
)

func init() {
	RegisterFileHandle("/file/stat", statHandler)
	RegisterFileHandle("/file/list", listHandler)
}

func statHandler(w http.ResponseWriter, r *http.Request) error {
	authPath, _, err := authWrite(r)
	if err != nil {
		log.Info(err)
		return writeMsg(w, 401, errors.As(err).Code())
	}

	fStat, err := os.Stat(authPath)
	if err != nil {
		if os.IsNotExist(err) {
			return writeMsg(w, 404, "filepath not exist")
		}
		return errors.As(err, authPath)
	}
	return writeJson(w, 200, &utils.ServerFileStat{
		FileName:    ".",
		IsDirFile:   fStat.IsDir(),
		FileSize:    fStat.Size(),
		FileModTime: fStat.ModTime(),
	})
}

func listHandler(w http.ResponseWriter, r *http.Request) error {
	authPath, _, err := authWrite(r)
	if err != nil {
		log.Info(err)
		return writeMsg(w, 401, errors.As(err).Code())
	}

	fStat, err := os.Stat(authPath)
	if err != nil {
		if os.IsNotExist(err) {
			return writeMsg(w, 404, "filepath not exist")
		}
		return errors.As(err, authPath)
	}
	if !fStat.IsDir() {
		return writeJson(w, 200, []utils.ServerFileStat{
			utils.ServerFileStat{
				FileName:    ".",
				IsDirFile:   false,
				FileSize:    fStat.Size(),
				FileModTime: fStat.ModTime(),
			},
		})
	}
	dirs, err := ioutil.ReadDir(authPath)
	if err != nil {
		return errors.As(err)
	}
	result := []utils.ServerFileStat{}
	for _, fs := range dirs {
		size := int64(0)
		if !fs.IsDir() {
			size = fs.Size()
		}
		result = append(result, utils.ServerFileStat{
			FileName:    fs.Name(),
			IsDirFile:   fs.IsDir(),
			FileSize:    size,
			FileModTime: fs.ModTime(),
		})
	}
	return writeJson(w, 200, result)
}
