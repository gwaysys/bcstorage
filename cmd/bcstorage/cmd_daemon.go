package main

import (
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gwaylib/errors"
	"github.com/gwaylib/log"
	"github.com/gwaysys/bcstorage/lib/cert"
	"github.com/urfave/cli/v2"
)

var daemonCmd = &cli.Command{
	Name:  "daemon",
	Usage: "Start a daemon to run storage server",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "root",
			Value: "/data/zfs", // default at current directory.
			Usage: "disk root path",
		},
		&cli.StringFlag{
			Name:  "repo",
			Value: "/data/zfs/.openz", // default at current directory.
			Usage: "storage work home in storage",
		},
		&cli.IntFlag{
			Name:  "ecc",
			Usage: "TODO:ecc nodes",
			Value: 0,
		},
		&cli.BoolFlag{
			Name:  "export-nfs",
			Usage: "export nfs service, depend on nfs-server",
			Value: false,
		},
	},
	Before: func(cctx *cli.Context) error {
		_rootPathFlag = cctx.String("root")
		_repoPathFlag = cctx.String("repo")
		return os.MkdirAll(_repoPathFlag, 0700)
	},
	Action: func(cctx *cli.Context) error {
		rootPath := _rootPathFlag
		repoPath := _repoPathFlag

		// stop the chattr protect.
		// TODO: remove this code
		if err := ioutil.WriteFile(filepath.Join(rootPath, "check.dat"), []byte("success"), 0600); err != nil {
			log.Fatal(errors.As(err, rootPath))
		}

		if err := InitLevelDB(filepath.Join(repoPath, _leveldbPath)); err != nil {
			panic(errors.As(err))
		}
		defer CloseLevelDB()
		initDaemonAuth()

		// export nfs for read
		// need install nfs-server
		// TODO: remove root export when userspace build done
		if cctx.Bool("export-nfs") {
			if err := ExportToNFS(rootPath); err != nil {
				return errors.As(err)
			}
		}

		// for http download and upload
		go func() {
			log.Infof("http-api:%s", _httpApiFlag)
			log.Fatal(http.ListenAndServe(_httpApiFlag, _fileHandler))
		}()

		// for https auth command
		crtPath := filepath.Join(_repoPathFlag, "storage_crt.pem")
		keyPath := filepath.Join(_repoPathFlag, "storage_key.pem")
		if err := cert.CreateTLSCert(crtPath, keyPath, []net.IP{}); err != nil {
			return errors.As(err)
		}
		log.Infof("auth-api:%s", _authApiFlag)
		return http.ListenAndServeTLS(_authApiFlag, crtPath, keyPath, _sysHandler)

	},
}
