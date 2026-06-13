package server

import (
	"net/http"

	"github.com/gwaylib/errors"
	"github.com/gwaylib/log"
	"github.com/gwaysys/bcstorage/module/bcrypt"
)

func init() {
	RegisterSysHandle("/sys/auth/add", addAuthHandler)
	RegisterSysHandle("/sys/auth/reset", resetAuthHandler)
	RegisterSysHandle("/sys/auth/change", changeAuthHandler)
}

func addAuthHandler(w http.ResponseWriter, r *http.Request) error {
	auth, err := authAdmin(r)
	if err != nil {
		return writeMsg(w, 401, errors.As(err).Code())
	}
	if auth.User != adminUser {
		return writeMsg(w, 401, "need admin user")
	}
	if bcrypt.BcryptMatch(adminDefaultPwd, auth.Passwd) {
		return writeMsg(w, 401, "need reset the admin default passwd")
	}
	spaceName := auth.User

	// checking space
	if _, ok := _userMap.GetSpace(spaceName); !ok {
		space := &UserSpace{
			Name: spaceName,
		}
		if err := _userMap.AddSpace(space); err != nil {
			return errors.As(err)
		}
	}

	passwd := r.FormValue("passwd")
	newAuth := &UserAuth{
		User:   r.FormValue("user"),
		Passwd: bcrypt.BcryptPwd(passwd),
		Spaces: map[string]bool{spaceName: true},
	}
	if err := _userMap.UpdateAuth(newAuth); err != nil {
		return errors.As(err)
	}
	log.Infof("add auth success from:%s,user:%s", r.RemoteAddr, newAuth.User)

	return writeMsg(w, 200, "success")
}
func resetAuthHandler(w http.ResponseWriter, r *http.Request) error {
	auth, err := authAdmin(r)
	if err != nil {
		return writeMsg(w, 401, errors.As(err).Code())
	}
	if auth.User != adminUser {
		return writeMsg(w, 401, "need admin user")
	}

	passwd := r.FormValue("passwd")
	userAuth, ok := _userMap.GetAuth(r.FormValue("user"))
	if !ok {
		return writeMsg(w, 401, "user not found")
	}
	userAuth.Passwd = bcrypt.BcryptPwd(passwd)
	if err := _userMap.UpdateAuth(userAuth); err != nil {
		return errors.As(err)
	}

	log.Infof("reset auth success from:%s,user:%s", r.RemoteAddr, userAuth.User)

	return writeMsg(w, 200, "success")
}

func changeAuthHandler(w http.ResponseWriter, r *http.Request) error {
	auth, err := authAdmin(r)
	if err != nil {
		return writeMsg(w, 401, errors.As(err).Code())
	}

	passwd := r.FormValue("passwd")
	auth.Passwd = bcrypt.BcryptPwd(passwd)
	if err := _userMap.UpdateAuth(auth); err != nil {
		return errors.As(err)
	}

	log.Infof("changed auth success from:%s", r.RemoteAddr)

	return writeMsg(w, 200, passwd)
}
