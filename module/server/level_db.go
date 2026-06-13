package server

import (
	"encoding/json/v2"
	"fmt"
	"path/filepath"

	"github.com/gwaylib/errors"
	"github.com/gwaysys/bcstorage/module/bcrypt"
	"github.com/syndtr/goleveldb/leveldb"
	lerrors "github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var (
	_levelDB *leveldb.DB
)

const (
	_leveldbPath = "leveldb"

	_leveldb_prefix_user  = "_user.%s"  // username
	_leveldb_prefix_space = "_space.%s" // spacename
	_leveldb_prefix_del   = "_del.%s%s" // spacename.uuid
	_leveldb_prefix_token = "_token.%s" // auth tokens
)

func CloseLevelDB() error {
	if _levelDB != nil {
		return _levelDB.Close()
	}
	return nil
}

func InitLevelDB(path string) error {
	if _levelDB != nil {
		return errors.New("level db has inited").As(path)
	}

	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		if !lerrors.IsCorrupted(err) {
			return errors.As(err)
		}
		db, err = leveldb.RecoverFile(path, nil)
		if err != nil {
			return errors.As(err)
		}
	}
	_levelDB = db
	return nil
}

// reset user passwd directly to level db
func ResetPasswdToLevelDB(cfgPath, user, passwd string) error {
	if err := InitLevelDB(filepath.Join(cfgPath, _leveldbPath)); err != nil {
		return errors.As(err)
	}
	defer CloseLevelDB()
	userAuth := UserAuth{}
	userKey := fmt.Sprintf(_leveldb_prefix_user, user)
	if err := GetLevelDB(userKey, &userAuth); err != nil {
		return errors.As(err)
	}
	userAuth.Passwd = bcrypt.BcryptPwd(passwd)
	if err := PutLevelDB(userKey, &userAuth); err != nil {
		return errors.As(err)
	}
	return nil
}

func PutLevelDB(key string, val interface{}) error {
	data, err := json.Marshal(val)
	if err != nil {
		return errors.As(err, key, val)
	}
	if err := _levelDB.Put([]byte(key), data, nil); err != nil {
		return errors.As(err)
	}
	return nil
}

// if key not found, return errors.ErrNoData
func GetLevelDB(key string, val interface{}) error {
	data, err := _levelDB.Get([]byte(key), nil)
	if err != nil {
		if leveldb.ErrNotFound == err {
			return errors.ErrNoData.As(key)
		}
		return errors.As(err, key)
	}
	if err := json.Unmarshal(data, val); err != nil {
		return errors.As(err, key, string(data))
	}
	return nil
}

func DelLevelDB(key string) error {
	if err := _levelDB.Delete([]byte(key), nil); err != nil {
		return errors.As(err, key)
	}
	return nil
}

func IterLevelDB(prefix string, cb func(key, val []byte) error) error {
	iter := _levelDB.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	defer iter.Release()
	for iter.Next() {
		if err := cb(iter.Key(), iter.Value()); err != nil {
			return errors.As(err)
		}
	}
	return errors.As(iter.Error())
}
