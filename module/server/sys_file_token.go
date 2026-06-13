package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gwaylib/errors"
)

func init() {
	RegisterSysHandle("/sys/file/token", tokenHandler)
}

func tokenHandler(w http.ResponseWriter, r *http.Request) error {
	auth, err := authAdmin(r)
	if err != nil {
		return writeMsg(w, 401, errors.As(err).Code())
	}
	space := auth.User
	if r.Method == "POST" {
		path := r.FormValue("path") // TODO: ../../filepath
		if len(path) == 0 {
			return writeMsg(w, 403, "params failed")
		}
		token := r.FormValue("token") // specified value, using random value if not set
		grant := r.FormValue("grant") // grant privilege, a for public to all user, others undefined
		expStr := r.FormValue("exp")  // minute, "" means no expire for readonly
		if len(token) == 0 {
			token = uuid.New().String()
		}

		// 当 kind="a" 时，不使用 exp 参数，不设置过期时间
		var expAt time.Time
		if expStr == "" {
			exp := 60
			expAt = time.Now().Add(time.Duration(exp) * time.Minute)
		} else {
			exp, err := strconv.ParseInt(expStr, 10, 64)
			if err != nil {
				exp = 60
			}
			if exp > 0 {
				expAt = time.Now().Add(time.Duration(exp) * time.Minute)
			}
		}

		if err := _fileHandler.AddToken(space, path, token, grant, expAt); err != nil {
			return writeMsg(w, 500, errors.As(err).Error())
		}
		return writeMsg(w, 200, token)
	}

	if r.Method == "DELETE" {
		token := r.FormValue("tk")
		if len(token) == 0 {
			return writeMsg(w, 403, "params failed")
		}
		authToken, err := _fileHandler.GetToken(token)
		if err != nil {
			if errors.ErrNoData.Equal(err) {
				return writeMsg(w, 200, "success")
			}
			return writeMsg(w, 500, err.Error())
		}

		if auth.User != "admin" && authToken._space != auth.User {
			return writeMsg(w, 401, "auth failed")
		}

		if err := _fileHandler.DeleteToken(token); err != nil {
			return writeMsg(w, 500, errors.As(err).Error())
		}
		return writeMsg(w, 200, "success")
	}
	return writeMsg(w, 403, "unsupport method")
}
