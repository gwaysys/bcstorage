package main

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gwaysys/bcstorage/lib/bcrypt"
	"github.com/gwaylib/errors"
)

const (
	_leveldbPath = "leveldb"
)

// empty of md5 is 'd41d8cd98f00b204e9800998ecf8427e'
const adminUser = "admin"
const adminDefaultPwd = "d41d8cd98f00b204e9800998ecf8427e"

var (
	_userMap = UserMap{
		Auth:  map[string]UserAuth{},
		Space: map[string]UserSpace{},
	}
)

func genPasswd() string {
	token := [16]byte(uuid.New())
	if time.Now().UnixNano()%2 == 0 {
		return fmt.Sprintf("%X", md5.Sum(token[:]))
	}
	return fmt.Sprintf("%x", md5.Sum(token[:]))
}

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

func (u UserAuth) HasSpace(space string) bool {
	u.spaceLk.Lock()
	defer u.spaceLk.Unlock()
	if u.Spaces == nil {
		return false
	}
	_, ok := u.Spaces[space]
	return ok
}

type UserMap struct {
	lk    sync.Mutex
	Auth  map[string]UserAuth
	Space map[string]UserSpace
}

func initDaemonAuth() {
	_, ok := _userMap.GetAuth(adminUser)
	if !ok {
		if err := _userMap.UpdateAuth(UserAuth{
			User:   adminUser,
			Passwd: bcrypt.BcryptPwd(adminDefaultPwd),
		}); err != nil {
			panic(err)
		}
	}
}

func (u *UserMap) GetAuth(user string) (UserAuth, bool) {
	u.lk.Lock()
	defer u.lk.Unlock()
	a, ok := u.Auth[user]
	if !ok {
		if err := ScanLevelDB(fmt.Sprintf(_leveldb_prefix_user, user), &a); err != nil {
			return UserAuth{}, false
		}
		u.Auth[user] = a
	}
	return a, true
}

func (u *UserMap) AddSpace(space UserSpace) error {
	u.lk.Lock()
	defer u.lk.Unlock()
	u.Space[space.Name] = space
	if err := os.MkdirAll(filepath.Join(_rootPathFlag, space.Name), 0755); err != nil {
		return errors.As(err)
	}
	return PutLevelDB(fmt.Sprintf(_leveldb_prefix_space, space.Name), &space)
}
func (u *UserMap) GetSpace(name string) (UserSpace, bool) {
	u.lk.Lock()
	defer u.lk.Unlock()
	space, ok := u.Space[name]
	if !ok {
		if err := ScanLevelDB(fmt.Sprintf(_leveldb_prefix_space, name), &space); err != nil {
			return UserSpace{}, false
		}
		u.Space[name] = space
	}
	return space, true
}
func (u *UserMap) SpacePath(spaceName string) (string, error) {
	space, ok := u.GetSpace(spaceName)
	if !ok {
		return "", errors.ErrNoData.As(spaceName)
	}
	return filepath.Join(_rootPathFlag, space.Name), nil
}
func (u *UserMap) AddSpaceUsed(name string, val int64) error {
	u.lk.Lock()
	defer u.lk.Unlock()
	space, ok := u.Space[name]
	if !ok {
		return errors.ErrNoData.As(name)
	}
	space.Used += val
	u.Space[name] = space
	return PutLevelDB(fmt.Sprintf(_leveldb_prefix_space, name), &space)
}

func (u *UserMap) UpdateAuth(auth UserAuth) error {
	u.lk.Lock()
	defer u.lk.Unlock()

	u.Auth[auth.User] = auth
	return PutLevelDB(fmt.Sprintf(_leveldb_prefix_user, auth.User), &auth)
}

func validHttpFilePath(spaceName, file string) (string, error) {
	space, ok := _userMap.GetSpace(spaceName)
	if !ok {
		return "", errors.New("space not found").As(spaceName)
	}
	rootPath := _rootPathFlag
	absPath, err := filepath.Abs(filepath.Join(rootPath, space.Name, file))
	if err != nil {
		return "", errors.New("can not locate the abs path").As(rootPath, space.Name, file)
	}
	if !strings.HasPrefix(absPath, filepath.Join(rootPath, space.Name)) {
		return "", errors.New("the file path out of user space")
	}
	return absPath, nil
}

func authWrite(r *http.Request) (string, FileToken, error) {
	space := r.FormValue("space")
	file := r.FormValue("file")
	token := r.FormValue("token")
	fAuth, ok := _fileHandler.VerifyToken(space, file, token)
	if !ok {
		return "", FileToken{}, errors.New("verify token failed").As(space, file, token)
	}
	absPath, err := validHttpFilePath(space, file)
	if err != nil {
		return "", FileToken{}, errors.As(err)
	}
	return absPath, fAuth, nil
}

func authAdmin(r *http.Request) (UserAuth, error) {
	// auth
	username, passwd, ok := r.BasicAuth()
	if !ok {
		return UserAuth{}, errors.New("auth not set")
	}
	auth, ok := _userMap.GetAuth(username)
	if !ok {
		return UserAuth{}, errors.New("auth failed").As(username)
	}
	if !bcrypt.BcryptMatch(passwd, auth.Passwd) {
		// TODO: limit the failed
		return UserAuth{}, errors.New("auth failed").As(username, passwd)
	}
	return auth, nil
}
