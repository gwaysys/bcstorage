package server

import (
	"encoding/json/v2"
	"net/http"

	"github.com/gwaylib/errors"
)

type HandleFunc func(w http.ResponseWriter, r *http.Request) error

func writeMsg(w http.ResponseWriter, code int, msg string) error {
	w.WriteHeader(code)
	if _, err := w.Write([]byte(msg)); err != nil {
		return errors.As(err)
	}
	return nil
}
func writeJson(w http.ResponseWriter, code int, obj interface{}) error {
	output, err := json.Marshal(obj)
	if err != nil {
		return errors.As(err)
	}

	w.WriteHeader(code)
	if _, err := w.Write(output); err != nil {
		return errors.As(err)
	}
	return nil
}
