package bcrypt

import (
	"github.com/jameskeane/bcrypt"
)

func BcryptPwdSalt(oriPwd string) (pwd, salt string) {
	salt, err := bcrypt.Salt(10)
	if err != nil {
		panic(err)
	}
	pwd, err = bcrypt.Hash(oriPwd, salt)
	if err != nil {
		panic(err)
	}
	return pwd, salt
}

func BcryptPwd(oriPwd string) string {
	pwd, _ := BcryptPwdSalt(oriPwd)
	return pwd
}

func BcryptMatch(oriPwd, encodePwd string) bool {
	return bcrypt.Match(oriPwd, encodePwd)
}
