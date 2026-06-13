package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/gwaylib/errors"
	"github.com/gwaysys/bcstorage/module/bcrypt"
)

// empty of md5 is 'd41d8cd98f00b204e9800998ecf8427e'
const adminUser = "admin"
const adminDefaultPwd = "d41d8cd98f00b204e9800998ecf8427e"

var (
	_userMap = UserMap{}
)

type UserSpace struct {
	Name    string
	Attr    int32 // TODO
	Size    int64 // TODO
	Used    int64 // TODO
	Private bool  // TODO
}

type UserAuth struct {
	User    string
	Passwd  string
	Spaces  map[string]bool
	spaceLk sync.Mutex
}

func (u *UserAuth) HasSpace(space string) bool {
	u.spaceLk.Lock()
	defer u.spaceLk.Unlock()
	if u.Spaces == nil {
		return false
	}
	_, ok := u.Spaces[space]
	return ok
}

type UserMap struct {
	dataRepo string
	lk       sync.Mutex
	auth     sync.Map // map[string]*UserAuth
	space    sync.Map // map[string]*UserSpace
}

func initDaemonAuth() error {
	_, ok := _userMap.GetAuth(adminUser)
	if !ok {
		if err := _userMap.UpdateAuth(&UserAuth{
			User:   adminUser,
			Passwd: bcrypt.BcryptPwd(adminDefaultPwd),
		}); err != nil {
			return errors.As(err)
		}
	}
	return nil
}

func (u *UserMap) GetAuth(user string) (*UserAuth, bool) {
	a, ok := u.auth.Load(user)
	if !ok {
		auth := &UserAuth{}
		if err := GetLevelDB(fmt.Sprintf(_leveldb_prefix_user, user), auth); err != nil {
			return auth, false
		}
		u.auth.Store(user, auth)
		return auth, true
	}
	return a.(*UserAuth), true
}

func (u *UserMap) AddSpace(space *UserSpace) error {
	if err := os.MkdirAll(filepath.Join(u.dataRepo, space.Name), 0755); err != nil {
		return errors.As(err)
	}
	if err := PutLevelDB(fmt.Sprintf(_leveldb_prefix_space, space.Name), &space); err != nil {
		return errors.As(err)
	}
	u.space.Store(space.Name, space)
	return nil
}
func (u *UserMap) GetSpace(name string) (*UserSpace, bool) {
	s, ok := u.space.Load(name)
	if !ok {
		space := &UserSpace{}
		if err := GetLevelDB(fmt.Sprintf(_leveldb_prefix_space, name), space); err != nil {
			return nil, false
		}
		u.space.Store(name, space)
		return space, true
	}
	return s.(*UserSpace), true
}
func (u *UserMap) SpacePath(spaceName string) (string, error) {
	space, ok := u.GetSpace(spaceName)
	if !ok {
		return "", errors.ErrNoData.As(spaceName)
	}
	return filepath.Join(u.dataRepo, space.Name), nil
}
func (u *UserMap) AddSpaceUsed(name string, val int64) error {
	u.lk.Lock()
	defer u.lk.Unlock()
	space, ok := u.GetSpace(name)
	if !ok {
		return errors.ErrNoData.As(name)
	}
	space.Used += val
	if err := PutLevelDB(fmt.Sprintf(_leveldb_prefix_space, name), space); err != nil {
		return errors.As(err, name, val)
	}
	u.space.Store(name, space)
	return nil
}

func (u *UserMap) UpdateAuth(auth *UserAuth) error {
	if err := PutLevelDB(fmt.Sprintf(_leveldb_prefix_user, auth.User), &auth); err != nil {
		return errors.As(err)
	}
	u.auth.Store(auth.User, auth)
	return nil
}

func authAdmin(r *http.Request) (*UserAuth, error) {
	// auth
	username, passwd, ok := r.BasicAuth()
	if !ok {
		return nil, errors.New("auth not set")
	}
	auth, ok := _userMap.GetAuth(username)
	if !ok {
		return nil, errors.New("auth failed").As(username)
	}
	if !bcrypt.BcryptMatch(passwd, auth.Passwd) {
		// TODO: limit the failed
		return nil, errors.New("auth failed").As(username, passwd)
	}
	return auth, nil
}
