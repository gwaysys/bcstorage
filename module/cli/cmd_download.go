package cli

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gwaylib/errors"
	"github.com/gwaylib/log"
	"github.com/urfave/cli/v2"

	"github.com/gwaysys/bcstorage/module/client"
)

var DownloadCmd = &cli.Command{
	Name:  "download",
	Usage: "[remote path] [local path]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "token",
			Usage: "auth token",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "mode",
			Usage: "read mode, support mode: 'http', 'tcp'. http mode support directory and resume download, tcp only support one file for debug fuse function.",
			Value: "http",
		},
		&cli.IntFlag{
			Name:  "read-size",
			Usage: "buffer size for read",
			Value: 1024 * 1024,
		},
	},
	Action: func(cctx *cli.Context) error {
		if !cctx.Args().Present() {
			return fmt.Errorf("arguments with [remote path] [local path]")
		}
		args := cctx.Args()
		if args.Len() != 2 {
			return fmt.Errorf("arguments with [remote path] [local path]")
		}
		remotePath := args.Get(0)
		localPath := args.Get(1)

		end := make(chan os.Signal, 2)
		switch cctx.String("mode") {
		case "http":
			go func() {
				// TODO: process the download
				ctx := cctx.Context
				log.Infof("start download: %s->%s", remotePath, localPath)
				startTime := time.Now()
				fc := client.NewHttpClient(_httpApiFlag, cctx.String("token"))
				if err := fc.Download(ctx, localPath, remotePath); err != nil {
					log.Exit(1, errors.As(err).Error())
					return
				}
				now := time.Now()
				log.Infof("end download: %s->%s, took:%s", remotePath, localPath, now.Sub(startTime))
				end <- os.Interrupt
			}()
		default:
			return fmt.Errorf("unknow mode '%s'", cctx.String("mode"))

		}
		// TODO: show the process
		signal.Notify(end, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
		<-end
		return nil
	},
}
