package server

import (
	"context"
	"net/http"
	"os/exec"
	"time"

	"github.com/gwaylib/errors"
)

func init() {
	RegisterSysHandle("/sys/check", checkHandler)
}

func checkHandler(w http.ResponseWriter, r *http.Request) error {
	_sysHandler.checkCacheLk.Lock()
	defer _sysHandler.checkCacheLk.Unlock()
	now := time.Now()
	if _sysHandler.checkCache != nil && now.Sub(_sysHandler.checkCache.createTime) < time.Minute {
		return writeMsg(w, 200, _sysHandler.checkCache.out)
	}

	// TODO: confirm the status tool
	output, err := exec.CommandContext(context.TODO(), "zpool", "status", "-x").CombinedOutput()
	if err != nil {
		return errors.As(err)
	}
	_sysHandler.checkCache = &CheckCache{
		out:        string(output),
		createTime: now,
	}
	return writeMsg(w, 200, string(output))
}
