package cli

import (
	"github.com/gwaysys/bcstorage/module/version"
	"github.com/urfave/cli/v2"
)

var (
	// common flag
	_authListenFlag = ""
	_authApiFlag    = ""
	_httpListenFlag = ""
	_httpApiFlag    = ""
)

func NewApp() *cli.App {
	commands := []*cli.Command{
		DaemonCmd,
		GrantCmd,
		DownloadCmd,
		UploadCmd,
		SysCmd,
	}
	return &cli.App{
		Name:    "bc-storage",
		Usage:   "simple net storage with http, tcp",
		Version: version.BuildVersion(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "auth-listen",
				Usage: "auth listen address",
				Value: "127.0.0.1:1330",
			},
			&cli.StringFlag{
				Name:  "auth-addr",
				Usage: "auth api address",
				Value: "https://127.0.0.1:1330",
			},
			&cli.StringFlag{
				Name:  "http-listen",
				Usage: "http transfer api address",
				Value: ":1331",
			},
			&cli.StringFlag{
				Name:  "http-addr",
				Usage: "http transfer api address",
				Value: "http://127.0.0.1:1331",
			},
			&cli.StringFlag{
				Name:  "pxfs-listen",
				Usage: "TODO: implement this",
				Value: "127.0.0.1:1332",
			},
			&cli.StringFlag{
				Name:  "pxfs-addr",
				Usage: "TODO: implement this",
				Value: "127.0.0.1:1332",
			},
			&cli.StringFlag{
				Name:  "user",
				Usage: "system admin user for operation, also is the namespace",
				Value: "admin",
			},
			&cli.StringFlag{
				Name:  "passwd",
				Usage: "passwd for system admin",
				Value: "d41d8cd98f00b204e9800998ecf8427e",
			},
		},
		Commands: commands,
		Before: func(cctx *cli.Context) error {
			_authListenFlag = cctx.String("auth-listen")
			_authApiFlag = cctx.String("auth-addr")
			_httpListenFlag = cctx.String("http-listen")
			_httpApiFlag = cctx.String("http-addr")
			return nil
		},
	}
}
