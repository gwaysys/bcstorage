package cli

import (
	"net"
	"path/filepath"

	"github.com/gwaylib/errors"
	"github.com/gwaylib/log"
	"github.com/gwaysys/bcstorage/module/cert"
	"github.com/gwaysys/bcstorage/module/server"
	"github.com/urfave/cli/v2"
)

var DaemonCmd = &cli.Command{
	Name:  "daemon",
	Usage: "Start a daemon to run storage server",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "data-path",
			Value: "./data", // default at current directory.
			Usage: "data path",
		},
		&cli.StringFlag{
			Name:  "cfg-path",
			Value: "./.bcstorage",
			Usage: "configeration path",
		},
		// &cli.IntFlag{
		// 	Name:  "ecc",
		// 	Usage: "TODO:ecc nodes",
		// 	Value: 0,
		// },
		&cli.BoolFlag{
			Name:  "export-nfs",
			Usage: "export nfs service, depend on nfs-server",
			Value: false,
		},
	},
	Action: func(cctx *cli.Context) error {
		dataPath := cctx.String("data-path")
		cfgPath := cctx.String("cfg-path")

		if err := server.InitAll(dataPath, cfgPath); err != nil {
			return errors.As(err)
		}
		defer server.CloseLevelDB()

		// export nfs for read
		// need install nfs-server
		// TODO: remove root export when userspace build done
		if cctx.Bool("export-nfs") {
			if err := server.ExportToNFS(dataPath); err != nil {
				return errors.As(err)
			}
		}

		// for http download and upload
		go func() {
			log.Infof("http-listen:%s", _httpListenFlag)
			log.Fatal(server.HttpFileListen(_httpListenFlag))
		}()

		// for https auth command
		// TODO: create root crt
		crtPath := filepath.Join(cfgPath, "storage_crt.pem")
		keyPath := filepath.Join(cfgPath, "storage_key.pem")
		if err := cert.CreateTLSCert(crtPath, keyPath, []net.IP{}); err != nil {
			return errors.As(err)
		}

		log.Infof("auth-listen:%s", _authListenFlag)
		return server.HttpsAuthListen(_authListenFlag, crtPath, keyPath)

	},
}
