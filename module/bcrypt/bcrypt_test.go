package bcrypt

import (
	"fmt"
	"testing"

	"github.com/jameskeane/bcrypt"
)

func TestBcryptPwd(t *testing.T) {
	oriPwd := "123"
	salt, _ := bcrypt.Salt(10)
	pwd, _ := bcrypt.Hash(oriPwd, salt)
	fmt.Println(salt, pwd)
}
