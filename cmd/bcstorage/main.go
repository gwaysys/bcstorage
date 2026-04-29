package main

import (
	"os"

	"github.com/gwaylib/log"
	"github.com/urfave/cli/v2"
)

func BuildVersion() string {
	return "v1.0.0"
}

var (
	// common flag
	_repoPathFlag = ""
	_authApiFlag  = ""
	_httpApiFlag  = ""

	// flag for server
	_rootPathFlag = ""
	_sysHandler   = &SysHandler{}
	_fileHandler  = &FileHandler{
		token: map[string]FileToken{},
	}
)

func main() {
	commands := []*cli.Command{
		daemonCmd,
		downloadCmd,
		uploadCmd,
		sysCmd,
	}
	app := &cli.App{
		Name:    "lotus-storage",
		Usage:   "storage server to replace nfs for lotus",
		Version: BuildVersion(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "addr-auth",
				Usage: "auth api address",
				Value: "127.0.0.1:1330",
			},
			&cli.StringFlag{
				Name:  "addr-http",
				Usage: "http transfer api address",
				Value: "127.0.0.1:1331",
			},
			&cli.StringFlag{
				Name:  "addr-pxfs",
				Usage: "TODO: implement this",
				Value: "127.0.0.1:1332",
			},
			&cli.StringFlag{
				Name:  "user",
				Usage: "system admin user for operation",
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
			_authApiFlag = cctx.String("addr-auth")
			_httpApiFlag = cctx.String("addr-http")
			return nil
		},
	}
	app.Setup()
	if err := app.Run(os.Args); err != nil {
		log.Warnf("%+v", err)
		return
	}
}
